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
	"github.com/trebent/zerologr"
)

const (
	defaultDuration    = time.Minute * 5
	defaultReportPath  = "report.yaml"
	defaultInteractive = false
)

var (
	// global flag vars.
	//nolint:gochecknoglobals // modified by flag parsing
	duration time.Duration

	// report.
	reportPath string //nolint:gochecknoglobals // modified by flag parsing

	// interactive enables the live TUI dashboard.
	//nolint:gochecknoglobals // modified by flag parsing
	interactive bool

	// rootCmd holds the cobra root command for Usage access.
	//nolint:gochecknoglobals // package-level command for Usage access
	rootCmd *cobra.Command

	// ErrStopping is returned when there was an error stopping traffic or modules.
	// It is used to wrap any errors from traffic or module stopping to allow
	// callers to check for this specific case.
	ErrStopping = errors.New("error stopping traffic or modules")
)

//nolint:gochecknoinits // sets up global logger at package load
func init() {
	zerologr.SetVFieldName("v")
	zerologr.Set(zerologr.New(&zerologr.Opts{Console: true}).WithName("global"))
}

// Usage prints the usage information for the Arbiter command line arguments.
func Usage() {
	_ = rootCmd.Usage()
}

// Run the Arbiter. Blocks until SIGINT, SIGTERM or when the test duration
// runs out (5 minute default).
func Run(modules module.Modules) error {
	if err := module.Validate(modules); err != nil {
		return err
	}

	// Reset to defaults on each Run call.
	duration = defaultDuration
	reportPath = defaultReportPath
	interactive = defaultInteractive

	// Root cmd that all subcommands are added to.
	rootCmd = &cobra.Command{
		Use:           "arbiter",
		Short:         "Arbiter load testing.",
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	cliCmd, fileCmd, err := buildRunnerCmds(modules)
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

func buildRunnerCmds(modules module.Modules) (*cobra.Command, *cobra.Command, error) {
	// The runner flagset is passed to cli and file commands that run tests.
	runnerFlagSet := buildRunnerFlagSet()

	cliCmd, err := cli.NewCommand(modules, run)
	if err != nil {
		return nil, nil, err
	}
	cliCmd.Flags().AddFlagSet(runnerFlagSet)

	runnerPreRunE := func(_ *cobra.Command, _ []string) error {
		if duration < 1*time.Second {
			return errors.New("duration must be at least 1 second")
		}

		if reportPath == "" {
			return errors.New("report path cannot be empty")
		}

		// err is fine since the file does not have to exist prior to the test ending.
		stat, err := os.Stat(reportPath) //nolint:govet // shad
		if err == nil && stat.IsDir() {
			return errors.New("report path cannot be a directory")
		}

		if interactive {
			// Suppress log output while the TUI is active to prevent interference
			// TODO: create file sink instead to be able to capture logs in interactive mode as well
			zerologr.Set(logr.Discard())
		} else {
			zerologr.Set(zerologr.New(&zerologr.Opts{Console: true, Caller: true}).WithName("arbiter"))
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

			return run(meta)
		},
	}
	fileCmd.Flags().AddFlagSet(runnerFlagSet)

	return cliCmd, fileCmd, nil
}

func buildRunnerFlagSet() *pflag.FlagSet {
	runnerFlagSet := &pflag.FlagSet{}
	runnerFlagSet.DurationVarP(
		&duration,
		"duration",
		"d",
		defaultDuration,
		"The duration of the test run, minimum 1 second.",
	)
	runnerFlagSet.StringVarP(
		&reportPath,
		"report-path",
		"r",
		defaultReportPath,
		"Path to the final report.",
	)
	runnerFlagSet.BoolVarP(
		&interactive,
		"interactive",
		"i",
		defaultInteractive,
		"Start in interactive TUI mode with a live progress bar and per-operation statistics.",
	)
	return runnerFlagSet
}

// Runs the input modules and starts generating traffic. Creates a traffic
// model based on the modules opertation settings. Aborts on SIGINT, SIGTERM
// or when the test duration runs out. Will immediately exit if any module
// returns an error from its call to Run().
func run(metadata module.Metadata) error {
	zerologr.Info("Starting modules")

	if err := startModules(metadata); err != nil {
		zerologr.Error(err, "Start failure")
		return err
	}
	zerologr.Info("All modules started")

	reporter := setupReporter(metadata)

	// Start signal interceptor for SIGINT and SIGTERM
	signalCtx, signalCancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer signalCancel()

	// Traffic context with a timeout of the test's >>> duration <<<
	timeoutCtx, timeoutCancel := context.WithTimeout(signalCtx, duration)
	defer timeoutCancel()
	zerologr.Info("Traffic will run for: " + duration.String())

	// The reporter runs in its own context to allow reporting to finalize separately from traffic and module
	// shutdown.
	reporterCtx, reporterCancel := context.WithCancel(context.Background())
	defer reporterCancel()
	reporter.Start(reporterCtx)

	// Run traffic.
	if err := traffic.Run(timeoutCtx, metadata, reporter); err != nil {
		zerologr.Error(err, "Failed to start traffic")
		return err
	}

	zerologr.Info("Awaiting completion (SIGINT/SIGTERM or duration timeout)")
	select {
	case <-signalCtx.Done():
		zerologr.Info("Got stop signal")
		// no need to call timeoutCancel() here since the traffic context is a child of the signal context,
		// so will be cancelled automatically.
		// timeoutCancel()
	case <-timeoutCtx.Done():
		zerologr.Info("Deadline exceeded")
		// Needed to terminate the parent context, in case other's are reliant on it.
		signalCancel()
	}

	// stopErr accumulates any errors from stopping traffic and modules, and finalising the report,
	// to be returned at the end of the function.
	var stopErr error
	if stopErr = traffic.Stop(); stopErr != nil {
		zerologr.Error(stopErr, "Error when stopping traffic")
		stopErr = fmt.Errorf("%w: traffic stop: %w", ErrStopping, stopErr)
	}

	// Now that traffic has been stopped, we can stop the reporter to allow it to finalise the report.
	reporterCancel()

	zerologr.Info("Stopping modules")
	for _, m := range metadata {
		if moduleStopErr := m.Stop(); moduleStopErr != nil {
			zerologr.Error(moduleStopErr, "Module stop reported an error", "module", m.Name())
			stopErr = errors.Join(stopErr, fmt.Errorf("module %s stop: %w", m.Name(), moduleStopErr))
		}
	}

	zerologr.Info("Finalising report")
	reporterStopErr := reporter.Finalise()
	if reporterStopErr != nil {
		zerologr.Error(reporterStopErr, "Error when finalising report")
		stopErr = errors.Join(stopErr, fmt.Errorf("reporter stop: %w", reporterStopErr))
	}

	return stopErr
}

// Starts the input modules and logs any errors.
func startModules(meta []*module.Meta) error {
	for _, m := range meta {
		zerologr.Info("Starting", "module", m.Name())
		if err := m.Run(); err != nil {
			return fmt.Errorf("failed to start module %s: %w", m.Name(), err)
		}
	}

	return nil
}

// Creates the reporter(s). In interactive mode a collection reporter is
// returned that fans out to both a YAML reporter and the live TUI reporter.
func setupReporter(metadata module.Metadata) report.Reporter {
	yamlR := yamlreport.New(&yamlreport.Opts{
		Path: reportPath,
	})

	if interactive {
		return collection.New(yamlR, interactivereport.New(metadata, duration))
	}

	return yamlR
}
