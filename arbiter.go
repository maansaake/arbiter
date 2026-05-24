// Package arbiter implements the orchestration between reporting,
// traffic scheduling and startup/shutdown procedures.
package arbiter

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-logr/logr"
	"github.com/maansaake/arbiter/pkg/module"
	"github.com/maansaake/arbiter/pkg/report"
	"github.com/maansaake/arbiter/pkg/report/collection"
	interactivereport "github.com/maansaake/arbiter/pkg/report/interactive"
	yamlreport "github.com/maansaake/arbiter/pkg/report/yaml"
	"github.com/maansaake/arbiter/pkg/subcommand/cli"
	"github.com/maansaake/arbiter/pkg/subcommand/file"
	"github.com/maansaake/arbiter/pkg/subcommand/gen"
	"github.com/maansaake/arbiter/pkg/traffic"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/trebent/envparser"
	"github.com/trebent/zerologr"
)

type (
	// abtr encapsulates the state of the arbiter.
	abtr struct {
		opts *Opts

		// duration is the test duration.
		duration time.Duration
		// reportPath is the file path to write the report to.
		reportPath string
		// interactive is set when an interactive TUI reporting is used.
		interactive bool
		// logger is used for info-level logging.
		logger logr.Logger
		// errorLogger is the logger used for error logs by the reporter.
		errorLogger logr.Logger
	}

	// Opts provide the arbiter with run-time options.
	Opts struct {
		// ErrorLogPath is set to a path where error logs will be written by the reporter. Defaults to error.log if not set.
		ErrorLogPath string
		// InfoLogPath is set to a path where info logs will be written by the reporter. Defaults to info.log if not set.
		InfoLogPath string
	}
)

const (
	defaultInfoLogPath  = "info.log"
	defaultErrorLogPath = "error.log"
	defaultDuration     = time.Minute * 5
	defaultReportPath   = "report.yaml"
	defaultInteractive  = false
)

// defaultOpts sets zero-value fields to their defaults.
func (o *Opts) defaultOpts() {
	if o.ErrorLogPath == "" {
		o.ErrorLogPath = defaultErrorLogPath
	}

	if o.InfoLogPath == "" {
		o.InfoLogPath = defaultInfoLogPath
	}
}

var (
	// rootCmd holds the cobra root command for Usage access.
	//nolint:gochecknoglobals // package-level command for Usage access
	rootCmd *cobra.Command

	// ErrStopping is returned when there was an error stopping traffic or modules.
	// It is used to wrap any errors from traffic or module stopping to allow
	// callers to check for this specific case.
	ErrStopping = errors.New("error stopping traffic or modules")

	//nolint:gochecknoglobals // package-level since env-var
	logVerbosity = envparser.Register(&envparser.Opts[int]{
		Name:  "ABTR_LOG_VERBOSITY",
		Desc:  "Set the log verbosity level for the Arbiter. Higher values are more verbose. Default is 0.",
		Value: 0,
	})

	//nolint:gochecknoglobals // package-level since env-var
	workerLimit = envparser.Register(&envparser.Opts[int]{
		Name:  "ABTR_WORKER_LIMIT",
		Desc:  "Set the maximum number of concurrent workers per workload. Default is 10.",
		Value: 10, //nolint:mnd // default value
	})
)

// Usage prints the usage information for the Arbiter command line arguments.
func Usage() {
	_ = rootCmd.Usage()
}

// Run the Arbiter. Tests block until SIGINT, SIGTERM or when the test duration
// runs out (5 minute default).
func Run(modules module.Modules, opts *Opts) error {
	// Parse is idempotent, so can be called in subcommands without issue.
	_ = envparser.Parse()
	if opts == nil {
		opts = &Opts{}
	}
	opts.defaultOpts()

	if err := module.Validate(modules); err != nil {
		return err
	}

	// Root cmd that all subcommands are added to.
	rootCmd = &cobra.Command{
		Use:           "arbiter",
		Short:         "Arbiter load testing.",
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	infoLogger, errorLogger, err := setupLoggers(opts, logVerbosity.Value())
	if err != nil {
		return fmt.Errorf("failed to build loggers: %w", err)
	}

	abtr := &abtr{
		opts:        opts,
		duration:    defaultDuration,
		reportPath:  defaultReportPath,
		interactive: defaultInteractive,
		logger:      infoLogger,
		errorLogger: errorLogger,
	}

	cliCmd, fileCmd, err := abtr.buildRunnerCmds(modules)
	if err != nil {
		return err
	}

	rootCmd.AddCommand(
		cliCmd,
		fileCmd,
		&cobra.Command{
			Use:   gen.FlagsetName,
			Short: "Generate a test model file.",
			RunE: func(_ *cobra.Command, args []string) error {
				return gen.Generate(args, modules)
			},
		},
	)

	return rootCmd.Execute()
}

// buildRunnerCmds builds the cli and file subcommands for running tests,
// which have a shared set of flags. The cli command runs tests based on
// CLI arguments, while the file command runs tests based on a test model file.
func (a *abtr) buildRunnerCmds(modules module.Modules) (*cobra.Command, *cobra.Command, error) {
	// The runner flagset is passed to cli and file commands that run tests.
	runnerFlagSet := a.buildRunnerFlagSet()

	cliCmd, err := cli.NewCommand(modules, func(m module.Metadata) error {
		return a.run(m)
	})
	if err != nil {
		return nil, nil, err
	}
	cliCmd.Flags().AddFlagSet(runnerFlagSet)

	runnerPreRunE := func(_ *cobra.Command, _ []string) error {
		if a.duration < 1*time.Second {
			return errors.New("duration must be at least 1 second")
		}

		if a.reportPath == "" {
			return errors.New("report path cannot be empty")
		}

		// err is fine since the file does not have to exist prior to the test ending.
		stat, err := os.Stat(a.reportPath) //nolint:govet // shad
		if err == nil && stat.IsDir() {
			return errors.New("report path cannot be a directory")
		}

		return nil
	}
	cliCmd.PreRunE = runnerPreRunE

	fileCmd := &cobra.Command{
		Use:     file.FlagsetName,
		Short:   "Run from a test model file.",
		PreRunE: runnerPreRunE,
		RunE: func(_ *cobra.Command, args []string) error {
			meta, err := file.Parse(args, modules) //nolint:govet // shad
			if err != nil {
				return err
			}

			return a.run(meta)
		},
	}
	fileCmd.Flags().AddFlagSet(runnerFlagSet)

	return cliCmd, fileCmd, nil
}

// buildRunnerFlagSet builds the flagset used by both the cli and file subcommands.
func (a *abtr) buildRunnerFlagSet() *pflag.FlagSet {
	runnerFlagSet := &pflag.FlagSet{}
	runnerFlagSet.DurationVarP(
		&a.duration,
		"duration",
		"d",
		defaultDuration,
		"The duration of the test run, minimum 1 second.",
	)
	runnerFlagSet.StringVarP(
		&a.reportPath,
		"report-path",
		"r",
		defaultReportPath,
		"Path to the final report.",
	)
	runnerFlagSet.BoolVarP(
		&a.interactive,
		"interactive",
		"i",
		defaultInteractive,
		"Start in interactive TUI mode with per-operation statistics in real time.",
	)
	return runnerFlagSet
}

func (a *abtr) run(metadata module.Metadata) error {
	a.logger.Info("Starting modules")

	if err := a.startModules(metadata); err != nil {
		a.logger.Error(err, "Start failure")
		return err
	}
	a.logger.Info("All modules started")

	// Start signal interceptor for SIGINT and SIGTERM
	signalCtx, signalCancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer signalCancel()

	reporter := a.setupReporter(
		metadata,
		signalCtx, signalCancel,
	)

	// Traffic context with a timeout of the test's >>> duration <<<
	timeoutCtx, timeoutCancel := context.WithTimeout(signalCtx, a.duration)
	defer timeoutCancel()
	a.logger.Info("Traffic will run for: " + a.duration.String())

	// The reporter runs in its own context to allow reporting to finalize separately from traffic and module
	// shutdown.
	reporterCtx, reporterCancel := context.WithCancel(context.Background())
	defer reporterCancel()
	reporter.Start(reporterCtx)

	sched := traffic.New(&traffic.Opts{
		Logger:      a.logger,
		WorkerLimit: workerLimit.Value(),
	})

	// Run traffic.
	if err := sched.Run(timeoutCtx, metadata, reporter); err != nil {
		reporter.ReportError(err) // Report is done in case of early traffic failure, to highlight issues in the TUI.
		a.logger.Error(err, "Failed to start traffic")
		return err
	}

	a.logger.Info("Awaiting completion (SIGINT/SIGTERM or duration timeout)")
	select {
	case <-signalCtx.Done():
		a.logger.Info("Got stop signal")
		// no need to call timeoutCancel() here since the traffic context is a child of the signal context,
		// so will be cancelled automatically.
		// timeoutCancel()
	case <-timeoutCtx.Done():
		a.logger.Info("Deadline exceeded")
		// Needed to terminate the parent context, in case other's are reliant on it.
		signalCancel()
	}

	// stopErr accumulates any errors from stopping traffic and modules, and finalising the report,
	// to be returned at the end of the function.
	var stopErr error
	if stopErr = sched.Stop(); stopErr != nil {
		a.logger.Error(stopErr, "Error when stopping traffic")
		stopErr = fmt.Errorf("%w: traffic stop: %w", ErrStopping, stopErr)
	}

	// Now that traffic has been stopped, we can stop the reporter to allow it to finalise the report.
	reporterCancel()

	a.logger.Info("Stopping modules")
	for _, m := range metadata {
		if moduleStopErr := m.Stop(); moduleStopErr != nil {
			a.logger.Error(moduleStopErr, "Module stop reported an error", "module", m.Name())
			stopErr = errors.Join(stopErr, fmt.Errorf("module %s stop: %w", m.Name(), moduleStopErr))
		}
	}

	a.logger.Info("Finalising report")
	reporterStopErr := reporter.Finalise()
	if reporterStopErr != nil {
		a.logger.Error(reporterStopErr, "Error when finalising report")
		stopErr = errors.Join(stopErr, fmt.Errorf("reporter stop: %w", reporterStopErr))
	}

	return stopErr
}

// startModules starts the input modules and logs any errors.
func (a *abtr) startModules(meta []*module.Meta) error {
	for _, m := range meta {
		a.logger.Info("Starting", "module", m.Name())
		if err := m.Run(); err != nil {
			return fmt.Errorf("failed to start module %s: %w", m.Name(), err)
		}
	}

	return nil
}

// setupReporter creates the reporter(s). In interactive mode a collection reporter is
// returned that fans out to both a YAML reporter and the live TUI reporter.
// trafficCancel is called by the interactive reporter when the user requests an early
// stop (e.g. Ctrl-C inside the TUI), triggering the same shutdown path as
// SIGINT/SIGTERM on the parent context. trafficCtx is used by the interactive
// reporter to monitor the traffic progression and display helpful messages in the TUI.
func (a *abtr) setupReporter(
	metadata module.Metadata,
	//nolint:revive // the traffic context is special and not releated to the function really
	trafficCtx context.Context, trafficCancel func(),
) report.Reporter {
	yamlR := yamlreport.New(&yamlreport.Opts{
		Path:        a.reportPath,
		Logger:      a.logger,
		ErrorLogger: a.errorLogger,
	})

	if a.interactive {
		return collection.New(
			yamlR,
			interactivereport.New(
				metadata,
				a.duration,
				trafficCtx,
				trafficCancel,
			),
		)
	}

	return yamlR
}

// setupLoggers initialises the info and error loggers from the provided options.
// It returns the info logger, the error logger, and any error encountered.
func setupLoggers(opts *Opts, verbosity int) (logr.Logger, logr.Logger, error) {
	errorFile, err := os.Create(opts.ErrorLogPath)
	if err != nil {
		return logr.Logger{}, logr.Logger{}, err
	}

	var infoLogger logr.Logger
	infoFile, err := os.Create(opts.InfoLogPath)
	if err != nil {
		return logr.Logger{}, logr.Logger{}, err
	}

	infoLogger = zerologr.New(&zerologr.Opts{
		Output:  infoFile,
		Console: true,
		Caller:  true,
		V:       verbosity,
	}).WithName("abtr")

	errorLogger := zerologr.New(&zerologr.Opts{
		Output:  errorFile,
		Console: true,
		Caller:  false,
		V:       verbosity,
	})

	return infoLogger, errorLogger, nil
}
