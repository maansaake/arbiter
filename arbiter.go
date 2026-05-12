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

	"github.com/maansaake/arbiter/pkg/module"
	"github.com/maansaake/arbiter/pkg/report"
	yamlreport "github.com/maansaake/arbiter/pkg/report/yaml"
	"github.com/maansaake/arbiter/pkg/subcommand/cli"
	"github.com/maansaake/arbiter/pkg/subcommand/file"
	"github.com/maansaake/arbiter/pkg/subcommand/gen"
	"github.com/maansaake/arbiter/pkg/traffic"
	"github.com/spf13/cobra"
	"github.com/trebent/zerologr"
)

const (
	durationDefault   = time.Minute * 5
	reportPathDefault = "report.yaml"
)

var (
	// global flag vars.
	//nolint:gochecknoglobals // modified by flag parsing
	duration = durationDefault

	// logger.
	//nolint:gochecknoglobals // glob log
	startLogger = zerologr.New(&zerologr.Opts{Console: true}).WithName("start")

	// report.
	reportPath = reportPathDefault //nolint:gochecknoglobals // modified by flag parsing

	// rootCmd holds the cobra root command for Usage access.
	//nolint:gochecknoglobals // package-level command for Usage access
	rootCmd *cobra.Command
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
	// TODO: change to support > 1 module
	if len(modules) != 1 {
		return fmt.Errorf("currently only 1 module is supported, got %d", len(modules))
	}

	if err := module.Validate(modules); err != nil {
		return err
	}

	// Reset to defaults on each Run call.
	duration = durationDefault
	reportPath = reportPathDefault

	rootCmd = &cobra.Command{
		Use:           "arbiter",
		Short:         "Arbiter load testing framework.",
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	rootCmd.PersistentFlags().
		DurationVarP(&duration, "duration", "d", durationDefault, "The duration of the test run, minimum 1 second.")
	rootCmd.PersistentFlags().
		StringVarP(&reportPath, "report-path", "r", reportPathDefault, "Path to the final report.")

	cliCmd, err := cli.NewCommand(modules, run)
	if err != nil {
		return err
	}

	preRunE := func(_ *cobra.Command, _ []string) error {
		if duration < 1*time.Second {
			return errors.New("duration must be at least 1 second")
		}

		if reportPath == "" {
			return errors.New("report path cannot be empty")
		}

		stat, err := os.Stat(reportPath) //nolint:govet // shad
		if err == nil && stat.IsDir() {
			return errors.New("report path cannot be a directory")
		}

		return nil
	}
	cliCmd.PreRunE = preRunE

	rootCmd.AddCommand(
		cliCmd,
		&cobra.Command{
			Use:   gen.FlagsetName,
			Short: "Generate a test model file.",
			RunE: func(_ *cobra.Command, args []string) error {
				return gen.Generate(args, modules)
			},
		},
		&cobra.Command{
			Use:     file.FlagsetName,
			Short:   "Run from a test model file.",
			PreRunE: preRunE,
			RunE: func(_ *cobra.Command, args []string) error {
				meta, err := file.Parse(args, modules) //nolint:govet // shad
				if err != nil {
					return err
				}

				return run(meta)
			},
		},
	)

	return rootCmd.Execute()
}

// Runs the input modules and starts generating traffic. Creates a traffic
// model based on the modules opertation settings. Aborts on SIGINT, SIGTERM
// or when the test duration runs out. Will immediately exit if any module
// returns an error from its call to Run().
func run(metadata module.Metadata) error {
	startLogger.Info("Starting modules")

	if err := startModules(metadata); err != nil {
		startLogger.Error(err, "Start failure")
		return err
	}
	startLogger.Info("All modules started")

	reporter := setupReporter()

	// Start traffic and monitor, with a timeout of: test >duration<
	background := context.Background()
	deadlineCtx, deadlineCancel := context.WithTimeout(background, duration)
	defer deadlineCancel()
	startLogger.Info("Traffic will run for: " + duration.String())

	// Separate the reporter context to allow for finishing reporting separately from stopping traffic.
	reporterCtx, reporterCancel := context.WithCancel(background)
	defer reporterCancel()
	reporter.Start(reporterCtx)

	if err := traffic.Run(deadlineCtx, metadata, reporter); err != nil {
		startLogger.Error(err, "Failed to start traffic")
		return err
	}

	// Start signal interceptor for SIGINT and SIGTERM
	signalCtx, signalCancel := signal.NotifyContext(background, syscall.SIGINT, syscall.SIGTERM)
	defer signalCancel()
	startLogger.Info("Awaiting stop signal")
	select {
	case <-signalCtx.Done():
		startLogger.Info("Got stop signal")
	case <-deadlineCtx.Done():
		startLogger.Info("Deadline exceeded")
	}
	deadlineCancel()
	signalCancel()

	startLogger = startLogger.WithName("stopping")

	var stopErr error
	stopErr = traffic.Stop()
	if stopErr != nil {
		startLogger.Error(stopErr, "Error when stopping traffic")
	}

	// Stop it here to allow the scheduler to report all before shutting down.
	reporterCancel()

	startLogger.Info("Stopping modules")
	for _, m := range metadata {
		if moduleStopErr := m.Stop(); moduleStopErr != nil {
			startLogger.Error(moduleStopErr, "Module stop reported an error", "module", m.Name())
			stopErr = errors.Join(stopErr, fmt.Errorf("module %s: %w", m.Name(), moduleStopErr))
		}
	}

	startLogger.Info("Finalising report")
	reporterStopErr := reporter.Finalise()
	if reporterStopErr != nil {
		startLogger.Error(reporterStopErr, "Error when finalising report")
		stopErr = errors.Join(stopErr, reporterStopErr)
	}

	return stopErr
}

// Starts the input modules and logs any errors.
func startModules(meta []*module.Meta) error {
	for _, m := range meta {
		startLogger.Info("Starting", "module", m.Name())
		if err := m.Run(); err != nil {
			return fmt.Errorf("failed to start module %s: %w", m.Name(), err)
		}
	}

	return nil
}

// Creates a new YAML reporter.
func setupReporter() report.Reporter {
	reporter := yamlreport.New(&yamlreport.Opts{
		Path: reportPath,
	})

	return reporter
}
