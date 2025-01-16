// The arbiter package implements the orchestration between monitoring,
// reporting, traffic scheduling and startup/shutdown procedures.
package arbiter

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"slices"
	"syscall"
	"time"

	"tres-bon.se/arbiter/pkg/module"
	"tres-bon.se/arbiter/pkg/monitor"
	"tres-bon.se/arbiter/pkg/report"
	yamlreport "tres-bon.se/arbiter/pkg/report/yaml"
	"tres-bon.se/arbiter/pkg/subcommand"
	"tres-bon.se/arbiter/pkg/subcommand/cli"
	"tres-bon.se/arbiter/pkg/subcommand/file"
	"tres-bon.se/arbiter/pkg/subcommand/gen"
	"tres-bon.se/arbiter/pkg/traffic"
	"tres-bon.se/arbiter/pkg/zerologr"
)

const (
	durationDefault           = time.Minute * 5
	externalPrometheusDefault = false
	metricAddrDefault         = ":8888"
	reportPathDefault         = "report.yaml"
)

var (
	// flagset
	flagset *flag.FlagSet

	// global flag vars
	duration           time.Duration = durationDefault
	externalPrometheus bool          = externalPrometheusDefault

	// subcommand parsing vars.
	subcommands = []string{cli.FlagsetName, gen.FlagsetName, file.FlagsetName}

	// logger
	startLogger = zerologr.New(&zerologr.Opts{Console: true, V: 10}).WithName("start")

	// metrics
	metricAddr = metricAddrDefault

	// report
	reportPath = reportPathDefault

	ErrNoSubcommand       = errors.New("no subcommand given")
	ErrSubcommandNotFound = errors.New("subcommand not found")
	ErrDurationTooShort   = errors.New("duration has to be minimum 30 seconds")
	ErrInvalidMetricAddr  = errors.New("invalid metric address")
)

func init() {
	zerologr.VFieldName = "v"
	zerologr.SetLogger(zerologr.New(&zerologr.Opts{V: 0, Console: true}).WithName("global"))
}

// Runs the Arbiter. Blocks until SIGINT, SIGTERM or when the test duration
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
		if meta, err := cli.Parse(subcommandIndex, modules); err != nil {
			return err
		} else {
			return run(meta)
		}
	case gen.FlagsetName:
		// Generate run file based on input modules.
		return gen.Generate(subcommandIndex, modules)
	case file.FlagsetName:
		// Parse run file information and continue to run block.
		if meta, err := file.Parse(subcommandIndex, modules); err != nil {
			return err
		} else {
			return run(meta)
		}
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
	monitor := setupMonitor(reporter, metadata)

	// Start traffic and monitor, with a deadline set to time.Now() + test duration
	background := context.Background()
	deadline := time.Now().Add(duration)
	deadlineCtx, deadlineStop := context.WithDeadline(background, deadline)
	startLogger.Info("traffic will run until", "deadline", deadline)

	reporterCtx, reporterCancel := context.WithCancel(background)
	reporter.Start(reporterCtx)

	if err := monitor.Start(deadlineCtx); err != nil {
		startLogger.Error(err, "failed to start the monitor")
		panic(err)
	}

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
	signalStop()
	deadlineStop()

	startLogger = startLogger.WithName("stopping")

	startLogger.Info("stopping traffic")
	err := traffic.AwaitStop()
	if err != nil {
		startLogger.Error(err, "error when stopping traffic")
	}

	err = monitor.Stop()
	if err != nil {
		startLogger.Error(err, "error when stopping monitor")
	}

	// Stop it here to allow the scheduler to report all before shutting down.
	reporterCancel()

	startLogger.Info("stopping modules")
	for _, m := range metadata {
		if err := m.Stop(); err != nil {
			startLogger.Error(err, "module stop reported an error", "module", m.Name())
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
	flagset.BoolVar(&externalPrometheus, "monitor.metric.external", externalPrometheus, "External Prometheus instance, disables internal metric ticker and creates a HTTP server for scraping.")
	flagset.StringVar(&metricAddr, "monitor.metric.external.addr", metricAddr, "Prometheus metric endpoint address.")
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

	if _, err := net.ResolveTCPAddr("tcp", metricAddr); err != nil {
		fmt.Fprint(flagset.Output(), ErrInvalidMetricAddr.Error()+"\n")
		parseErrs = append(parseErrs, ErrInvalidMetricAddr)
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

// Creates a new monitor.
func setupMonitor(reporter report.Reporter, metadata subcommand.Metadata) *monitor.Monitor {
	m := monitor.New(metadata.MonitorOpts()...)
	m.Reporter = reporter
	m.ExternalPrometheus = externalPrometheus
	m.MetricAddr = metricAddr

	return m
}
