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
	"github.com/maansaake/arbiter/pkg/subcommand"
	"github.com/maansaake/arbiter/pkg/subcommand/cli"
	"github.com/maansaake/arbiter/pkg/subcommand/file"
	"github.com/maansaake/arbiter/pkg/subcommand/gen"
	"github.com/maansaake/arbiter/pkg/traffic"
	"github.com/maansaake/arbiter/pkg/zerologr"
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
	startLogger = zerologr.New(&zerologr.Opts{Console: true}).
			WithName("start")

	// report.
	reportPath = reportPathDefault //nolint:gochecknoglobals // modified by flag parsing

	ErrNoSubcommand       = errors.New("no subcommand given")
	ErrSubcommandNotFound = errors.New("subcommand not found")
	ErrDurationTooShort   = errors.New("duration has to be minimum 30 seconds")
)

func init() { //nolint:gochecknoinits // sets up global logger at package load
	zerologr.SetVFieldName("v")
	zerologr.SetLogger(zerologr.New(&zerologr.Opts{V: 0, Console: true}).WithName("global"))
}

// Run the Arbiter. Blocks until SIGINT, SIGTERM or when the test duration
// runs out (5 minute default).
func Run(modules module.Modules) error {
	// TODO: change to support > 1 module
	if len(modules) != 1 {
		panic("number of modules must be exactly one")
	}
	if err := module.Validate(modules); err != nil {
		return err
	}

	// Verify input arguments and parse subcommand.
	subcommandIndex, parseErrs := parseArguments(os.Args)
	if len(parseErrs) > 0 {
		flagset.SetOutput(os.Stderr)
		flagset.Usage()
		return errors.Join(parseErrs...)
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
		flagset.SetOutput(os.Stderr)
		err := fmt.Errorf("%w: %v", ErrSubcommandNotFound, os.Args)
		fmt.Fprint(flagset.Output(), err.Error()+"\n")
		flagset.Usage()

		return err
	}
}

// Runs the input modules and starts generating traffic. Creates a traffic
// model based on the modules opertation settings. Aborts on SIGINT, SIGTERM
// or when the test duration runs out. Will immediately exit if any module
// returns an error from its call to Run().
func run(metadata subcommand.Metadata) error {
	startLogger.Info("preparing to run the modules")

	if err := startModules(metadata); err != nil {
		startLogger.Error(err, "start failure")
		return err
	}
	startLogger.Info("all modules started")

	reporter := setupReporter()

	// Start traffic and monitor, with a deadline set to time.Now() + test duration
	background := context.Background()
	deadline := time.Now().Add(duration)
	deadlineCtx, deadlineStop := context.WithDeadline(background, deadline)
	startLogger.Info("traffic will run until", "deadline", deadline)

	reporterCtx, reporterCancel := context.WithCancel(background)
	reporter.Start(reporterCtx)

	if err := traffic.Run(deadlineCtx, metadata, reporter); err != nil {
		startLogger.Error(err, "failed to start traffic")
		panic(err)
	}

	// Start signal interceptor for SIGINT and SIGTERM
	signalCtx, signalStop := signal.NotifyContext(background, syscall.SIGINT, syscall.SIGTERM)
	startLogger.Info("awaiting stop signal")
	select {
	case <-signalCtx.Done():
		startLogger.Info("got stop signal")
	case <-deadlineCtx.Done():
		startLogger.Info("deadline exceeded")
	}
	deadlineStop()
	signalStop()

	startLogger = startLogger.WithName("stopping")

	err := traffic.Stop()
	if err != nil {
		startLogger.Error(err, "error when stopping traffic")
	}

	// Stop it here to allow the scheduler to report all before shutting down.
	reporterCancel()

	startLogger.Info("stopping modules")
	for _, m := range metadata {
		if stopErr := m.Stop(); stopErr != nil {
			startLogger.Error(stopErr, "module stop reported an error", "module", m.Name())
		}
	}

	startLogger.Info("finalising report")
	return reporter.Finalise()
}

// Parses the input arguments and returns the index of the subcommand and any
// parsing errors.
func parseArguments(args []string) (int, []error) {
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

	// Ingore error since we're using ExitOnError.
	_ = flagset.Parse(os.Args[1:])

	var parseErrs []error
	subcommandIndex := slices.IndexFunc(args, func(arg string) bool {
		return slices.Contains(subcommands, arg)
	})

	if subcommandIndex == -1 {
		fmt.Fprint(flagset.Output(), ErrNoSubcommand.Error()+"\n")
		parseErrs = append(parseErrs, ErrNoSubcommand)
	}

	if duration < 30*time.Second {
		fmt.Fprint(flagset.Output(), ErrDurationTooShort.Error()+"\n")
		parseErrs = append(parseErrs, ErrDurationTooShort)
	}

	return subcommandIndex, parseErrs
}

// Starts the input modules and logs any errors.
func startModules(meta []*subcommand.Meta) error {
	for _, m := range meta {
		startLogger.Info("starting", "module", m.Name())
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
