// The arbiter package implements the orchestration between monitoring,
// reporting, traffic scheduling and startup/shutdown procedures.
package arbiter

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"slices"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"tres-bon.se/arbiter/pkg/arg"
	"tres-bon.se/arbiter/pkg/module"
	"tres-bon.se/arbiter/pkg/monitor"
	"tres-bon.se/arbiter/pkg/monitor/cpu"
	"tres-bon.se/arbiter/pkg/monitor/log"
	"tres-bon.se/arbiter/pkg/monitor/memory"
	"tres-bon.se/arbiter/pkg/monitor/metric"
	"tres-bon.se/arbiter/pkg/report"
	"tres-bon.se/arbiter/pkg/subcommand"
	"tres-bon.se/arbiter/pkg/subcommand/cli"
	"tres-bon.se/arbiter/pkg/subcommand/file"
	"tres-bon.se/arbiter/pkg/subcommand/gen"
	"tres-bon.se/arbiter/pkg/traffic"
	"tres-bon.se/arbiter/pkg/zerologr"
)

const (
	MONITOR_DISABLE_METRIC_TICKER_DEFAULT = false
	DISABLE_ARBITER_METRIC_SERVER_DEFAULT = false
)

var (
	// global flag vars.
	duration                   time.Duration = time.Minute * 5
	monitorDisableMetricTicker bool          = MONITOR_DISABLE_METRIC_TICKER_DEFAULT
	disableMetricServer        bool          = DISABLE_ARBITER_METRIC_SERVER_DEFAULT

	// subcommand parsing vars.
	subcommands     = []string{arg.FLAGSET, gen.FLAGSET, file.FLAGSET}
	subcommandIndex = -1

	// log.
	startLogger = zerologr.New(&zerologr.Opts{Console: true, V: 10}).WithName("start")

	// metrics.
	metricAddr = ":8888"
	// metricPrefix = "arbiter"

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
	flag.BoolVar(&monitorDisableMetricTicker, "monitor.metric.disable.ticker", monitorDisableMetricTicker, "Disable the monitor metric ticker.")
	flag.BoolVar(&disableMetricServer, "disable.metric.server", disableMetricServer, "Disable the arbiter metric server.")

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
		// Register module arguments and continue to run block.
		if meta, err := cli.Register(subcommandIndex, modules); err != nil {
			return err
		} else {
			return run(meta)
		}
	case gen.FLAGSET:
		// Generate run file based on input module.
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
func run(meta []*subcommand.ModuleMeta) error {
	startLogger.Info("preparing to run the modules")

	if err := startModules(meta); err != nil {
		startLogger.Error(err, "start failure")
		return err
	}
	startLogger.Info("all modules started")

	reporter := setupReporter()
	monitor := setupMonitor(reporter, meta)
	metricServer := setupMetricServer(monitor, meta)

	// Start traffic and monitor, with a deadline set to time.Now() + test duration
	background := context.Background()
	deadline := time.Now().Add(duration)
	deadlineCtx, deadlineStop := context.WithDeadline(background, deadline)
	startLogger.Info("traffic will run until", "deadline", deadline)

	if err := monitor.Start(deadlineCtx); err != nil {
		startLogger.Error(err, "failed to start the monitor")
		panic(err)
	}

	if err := traffic.Run(deadlineCtx, meta, reporter); err != nil {
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
	stopCtx, stopCancel := context.WithDeadline(context.Background(), time.Now().Add(5*time.Second))
	defer stopCancel()

	if metricServer != nil {
		startLogger.Info("stopping metric server")
		if err := metricServer.Shutdown(stopCtx); err != nil {
			startLogger.Error(err, "metric server shutdown error")
		}
	}

	startLogger.Info("stopping traffic")
	traffic.AwaitStop()

	startLogger.Info("stopping modules")
	for _, m := range meta {
		if err := m.Stop(); err != nil {
			startLogger.Error(err, "module stop reported an error", "module", m.Name())
		}
	}

	startLogger.Info("finalising report")
	reporter.Finalise()

	return nil
}

func startModules(meta []*subcommand.ModuleMeta) error {
	for _, m := range meta {
		startLogger.Info("starting", "module", m.Name())
		if err := m.Run(); err != nil {
			return fmt.Errorf("failed to start module %s: %w", m.Name(), err)
		}
	}
	return nil
}

func setupReporter() report.Reporter {
	return &report.YAMLReporter{}
}

func setupMonitor(reporter report.Reporter, meta []*subcommand.ModuleMeta) *monitor.Monitor {
	monitor := &monitor.Monitor{
		Reporter:            reporter,
		DisableMetricTicker: monitorDisableMetricTicker,
	}

	for _, m := range meta {
		if m.PID != cli.NO_PERFORMANCE_PID {
			//nolint:gosec
			monitor.CPU = cpu.NewLocalCPUMonitor(int32(m.PID))
			//nolint:gosec
			monitor.Memory = memory.NewLocalMemoryMonitor(int32(m.PID))
		}

		if m.LogFile != cli.NO_LOG_FILE {
			monitor.Log = log.NewLogFileMonitor(m.LogFile)
		}

		if m.MetricEndpoint != cli.NO_METRIC_ENDPOINT {
			// TODO: metric endpoints should be per test module.
			monitor.Metric = metric.NewMetricMonitor(m.MetricEndpoint)
		}
	}

	return monitor
}

func setupMetricServer(monitor *monitor.Monitor, meta []*subcommand.ModuleMeta) *http.Server {
	// Start metric server
	var metricServer *http.Server

	if !disableMetricServer {
		go func() {
			//nolint:gosec
			metricServer = &http.Server{
				Addr:    metricAddr,
				Handler: http.DefaultServeMux, // Use the default handler
			}

			// Arbiter metrics
			http.Handle("/metrics", promhttp.Handler())

			// If metric endpoint(s) registered
			for _, m := range meta {
				if m.MetricEndpoint != cli.NO_METRIC_ENDPOINT {
					http.HandleFunc(fmt.Sprintf("/metrics-%s", m.Name()), func(w http.ResponseWriter, r *http.Request) {
						bs, err := monitor.PullMetrics()
						if err != nil {
							w.WriteHeader(500)
						}
						_, err = w.Write(bs)
						if err != nil {
							w.WriteHeader(500)
						}
					})
				}
			}

			startLogger.Info("running metrics server on", "address", metricAddr)
			if err := metricServer.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
				startLogger.Error(err, "unexpected error")
			} else {
				startLogger.Info("metrics server shut down")
			}
		}()
	}

	return metricServer
}
