// Package arbiter implements the orchestration between reporting,
// traffic scheduling and startup/shutdown procedures.
package arbiter

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"slices"
	"syscall"
	"time"

	"github.com/maansaake/arbiter/pkg/module"
	"github.com/maansaake/arbiter/pkg/report"
	yamlreport "github.com/maansaake/arbiter/pkg/report/yaml"
	"github.com/maansaake/arbiter/pkg/subcommand/cli"
	"github.com/maansaake/arbiter/pkg/subcommand/file"
	"github.com/maansaake/arbiter/pkg/subcommand/gen"
	"github.com/maansaake/arbiter/pkg/traffic"
	"github.com/trebent/zerologr"
)

const (
	durationDefault   = time.Minute * 5
	reportPathDefault = "report.yaml"
)

var (
	// flagset.
	//nolint:gochecknoglobals // package-level flagset for CLI parsing
	flagset *flag.FlagSet

	// global flag vars.
	//nolint:gochecknoglobals // modified by flag parsing
	duration = durationDefault

	// subcommand parsing vars.
	//nolint:gochecknoglobals // modified by flag parsing
	subcommands = []string{
		cli.FlagsetName,
		gen.FlagsetName,
		file.FlagsetName,
	}

	// logger.
	//nolint:gochecknoglobals // glob log
	startLogger = zerologr.New(&zerologr.Opts{Console: true}).WithName("start")

	// report.
	reportPath = reportPathDefault //nolint:gochecknoglobals // modified by flag parsing

	ErrNoSubcommand       = errors.New("no subcommand given")
	ErrSubcommandNotFound = errors.New("subcommand not found")
	ErrDurationTooShort   = errors.New("duration cannot be less than 1 second")
	ErrParsingFailed      = errors.New("parsing failed")
)

//nolint:gochecknoinits // sets up global logger at package load
func init() {
	zerologr.SetVFieldName("v")
	zerologr.Set(zerologr.New(&zerologr.Opts{Console: true}).WithName("global"))
}

// Usage prints the usage information for the Arbiter command line arguments.
func Usage() {
	flagset.Usage()
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

	// Verify input arguments and parse subcommand.
	subcommandIndex, parseErr := parseArguments(os.Args)
	if parseErr != nil {
		return parseErr
	}

	// Check invoked subcommand
	switch os.Args[subcommandIndex] {
	case cli.FlagsetName:
		// Parse module arguments and continue to run block.
		meta, err := cli.Parse(subcommandIndex, modules)
		if err != nil {
			return err
		}

		return run(meta)
	case gen.FlagsetName:
		// Generate run file based on input modules.
		return gen.Generate(subcommandIndex, modules)
	case file.FlagsetName:
		// Parse run file information and continue to run block.
		meta, err := file.Parse(subcommandIndex, modules)
		if err != nil {
			return err
		}

		return run(meta)
	default:
		return fmt.Errorf("%w: %v", ErrSubcommandNotFound, os.Args)
	}
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

	// Start traffic and monitor, with a deadline set to time.Now() + test duration
	background := context.Background()
	deadline := time.Now().Add(duration)
	deadlineCtx, deadlineCancel := context.WithDeadline(background, deadline)
	defer deadlineCancel()
	startLogger.Info("Traffic will run until: " + deadline.String())

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

// Parses the input arguments and returns the index of the subcommand and any
// parsing errors.
func parseArguments(args []string) (int, error) {
	flagset = flag.NewFlagSet("arbiter", flag.ExitOnError)

	formatFlagset := func(fset string) string {
		return fmt.Sprintf("%-10s", fset)
	}

	flagset.SetOutput(os.Stdout)
	flagset.Usage = func() {
		fmt.Fprintf(flagset.Output(), "%s [subcommand]\n\n", os.Args[0])
		fmt.Fprint(flagset.Output(), "subcommands:\n")
		fmt.Fprintf(flagset.Output(), "  %s Run using CLI flags.\n", formatFlagset(cli.FlagsetName))
		fmt.Fprintf(flagset.Output(), "  %s Generate a test model file.\n", formatFlagset(gen.FlagsetName))
		fmt.Fprintf(flagset.Output(), "  %s Run from a test model file.\n", formatFlagset(file.FlagsetName))
		fmt.Fprint(flagset.Output(), "\n")
		fmt.Fprint(flagset.Output(), "global flags:\n")
		flagset.PrintDefaults()
	}

	// Global flags
	flagset.DurationVar(&duration, "duration", duration, "The duration of the test run, minimum 30 seconds.")
	flagset.StringVar(&reportPath, "report.path", reportPath, "Path to the final report.")

	var totalErr error
	parseErr := flagset.Parse(os.Args[1:])
	if parseErr != nil {
		totalErr = fmt.Errorf("%w: %w", ErrParsingFailed, parseErr)
	}

	subcommandIndex := slices.IndexFunc(args, func(arg string) bool {
		return slices.Contains(subcommands, arg)
	})

	if subcommandIndex == -1 {
		flagset.Usage()
		os.Exit(1)
	}

	if duration < 1*time.Second {
		totalErr = fmt.Errorf("%w: %w", totalErr, ErrDurationTooShort)
	}

	return subcommandIndex, totalErr
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
