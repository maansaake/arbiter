// The arbiter package implements the orchestration between monitoring,
// reporting, traffic scheduling and startup/shutdown procedures.
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

	"tres-bon.se/arbiter/pkg/arg"
	"tres-bon.se/arbiter/pkg/module"
	"tres-bon.se/arbiter/pkg/monitor"
	"tres-bon.se/arbiter/pkg/report"
	"tres-bon.se/arbiter/pkg/subcommand"
	"tres-bon.se/arbiter/pkg/subcommand/cli"
	"tres-bon.se/arbiter/pkg/subcommand/file"
	"tres-bon.se/arbiter/pkg/subcommand/gen"
	"tres-bon.se/arbiter/pkg/traffic"
	"tres-bon.se/arbiter/pkg/zerologr"
)

const (
	DURATION_DEFAULT            = time.Minute * 5
	EXTERNAL_PROMETHEUS_DEFAULT = false
	METRIC_ADDR_DEFAULT         = ":8888"
	REPORT_PATH_DEFAULT         = "report.yaml"
)

var (
	// global flag vars.
	duration           time.Duration = DURATION_DEFAULT
	externalPrometheus bool          = EXTERNAL_PROMETHEUS_DEFAULT

	// subcommand parsing vars.
	subcommands     = []string{arg.FLAGSET, gen.FLAGSET, file.FLAGSET}
	subcommandIndex = -1

	// log.
	startLogger = zerologr.New(&zerologr.Opts{Console: true, V: 10}).WithName("start")

	// metrics.
	metricAddr = METRIC_ADDR_DEFAULT

	// report
	reportPath = REPORT_PATH_DEFAULT

	ErrParseError         = errors.New("flag parse error")
	ErrNoSubcommand       = errors.New("no subcommand given")
	ErrSubcommandNotFound = errors.New("subcommand not found")
	ErrDurationTooShort   = errors.New("duration has to be minimum 30 seconds")
)

func init() {
	zerologr.SetLogger(zerologr.New(&zerologr.Opts{Console: true}).WithName("global"))
}

// Runs the Arbiter. Blocks until SIGINT, SIGTERM or when the test duration
// runs out (5 minute default).
func Run(modules module.Modules) error {
	formatFlagset := func(fset string) string {
		return fmt.Sprintf("%-10s", fset)
	}

	// TODO: change to support > 1 module
	if len(modules) != 1 {
		panic("number of modules must be exactly one")
	}
	if err := module.Validate(modules); err != nil {
		return err
	}

	flag.CommandLine.SetOutput(os.Stdout)
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "%s [subcommand]\n\n", os.Args[0])
		fmt.Fprint(flag.CommandLine.Output(), "subcommands:\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  %s Run using CLI flags.\n", formatFlagset(arg.FLAGSET))
		fmt.Fprintf(flag.CommandLine.Output(), "  %s Generate a test model file.\n", formatFlagset(gen.FLAGSET))
		fmt.Fprintf(flag.CommandLine.Output(), "  %s Run from a test model file.\n", formatFlagset(file.FLAGSET))
		fmt.Fprint(flag.CommandLine.Output(), "\n")
		fmt.Fprint(flag.CommandLine.Output(), "global flags:\n")
		flag.PrintDefaults()
	}

	// Global flags
	flag.DurationVar(&duration, "duration", duration, "The duration of the test run, minimum 30 seconds.")
	flag.BoolVar(&externalPrometheus, "monitor.metric.external", externalPrometheus, "External Prometheus instance, disables internal metric ticker and creates a HTTP server for scraping.")
	flag.StringVar(&metricAddr, "monitor.metric.external.addr", metricAddr, "Prometheus metric endpoint address.")
	flag.StringVar(&reportPath, "report.path", reportPath, "Path to the final report.")

	// To trigger on --help and parse global flags
	flag.Parse()

	// Verify a subcommand has been invoked.
	subcommandIndex = slices.IndexFunc(os.Args, func(e string) bool {
		return slices.Contains(subcommands, e)
	})
	parseErrs := []error{}
	if subcommandIndex == -1 {
		fmt.Fprint(flag.CommandLine.Output(), ErrNoSubcommand.Error()+"\n")
		parseErrs = append(parseErrs, ErrNoSubcommand)
	}

	if duration < 30*time.Second {
		fmt.Fprint(flag.CommandLine.Output(), ErrDurationTooShort.Error()+"\n")
		parseErrs = append(parseErrs, ErrDurationTooShort)
	}

	if len(parseErrs) > 1 {
		flag.CommandLine.SetOutput(os.Stderr)
		flag.Usage()
		return errors.Join(parseErrs...)
	}

	// Check invoked subcommand
	switch os.Args[subcommandIndex] {
	case arg.FLAGSET:
		// Parse module arguments and continue to run block.
		if meta, err := cli.Parse(subcommandIndex, modules); err != nil {
			return err
		} else {
			return run(meta)
		}
	case gen.FLAGSET:
		// Generate run file based on input modules.
		return gen.Generate(subcommandIndex, modules)
	case file.FLAGSET:
		// Parse run file information and continue to run block.
		if meta, err := file.Parse(subcommandIndex, modules); err != nil {
			return err
		} else {
			return run(meta)
		}
	default:
		flag.CommandLine.SetOutput(os.Stderr)
		err := fmt.Errorf("%w: %v", ErrSubcommandNotFound, os.Args)
		fmt.Fprint(flag.CommandLine.Output(), err.Error()+"\n")
		flag.Usage()
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
	traffic.AwaitStop()

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

func startModules(meta []*subcommand.Meta) error {
	for _, m := range meta {
		startLogger.Info("starting", "module", m.Name())
		if err := m.Run(); err != nil {
			return fmt.Errorf("failed to start module %s: %w", m.Name(), err)
		}
	}
	return nil
}

func setupReporter() report.Reporter {
	reporter := report.NewYAML(&report.YAMLOpts{
		Path: reportPath,
	})

	return reporter
}

func setupMonitor(reporter report.Reporter, metadata subcommand.Metadata) *monitor.Monitor {
	m := monitor.New(metadata.MonitorOpts()...)
	m.Reporter = reporter
	m.ExternalPrometheus = externalPrometheus
	m.MetricAddr = metricAddr

	return m
}
