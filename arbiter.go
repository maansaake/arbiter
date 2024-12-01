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
	"tres-bon.se/arbiter/pkg/module"
	"tres-bon.se/arbiter/pkg/monitor"
	"tres-bon.se/arbiter/pkg/monitor/cpu"
	"tres-bon.se/arbiter/pkg/monitor/log"
	"tres-bon.se/arbiter/pkg/monitor/memory"
	"tres-bon.se/arbiter/pkg/monitor/metric"
	"tres-bon.se/arbiter/pkg/report"
	"tres-bon.se/arbiter/pkg/subcommand/cli"
	"tres-bon.se/arbiter/pkg/subcommand/file"
	"tres-bon.se/arbiter/pkg/subcommand/gen"
	"tres-bon.se/arbiter/pkg/traffic"
	"tres-bon.se/arbiter/pkg/zerologr"
)

const (
	MONITOR_DISABLE_METRIC_TICKER_DEFAULT = false
	DISABLE_ARBITER_METRIC_SERVER         = false
)

var (
	// global flag vars.
	duration                   time.Duration = time.Minute * 5
	monitorDisableMetricTicker bool          = MONITOR_DISABLE_METRIC_TICKER_DEFAULT
	disableMetricServer        bool          = DISABLE_ARBITER_METRIC_SERVER

	// subcommand parsing vars.
	subcommands     = []string{cli.FLAGSET, gen.FLAGSET, file.FLAGSET}
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

// Runs the Arbiter. Blocks until SIGINT, SIGTERM or when the test duration
// runs out (5 minute default).
func Run(modules module.Modules) error {
	formatFlagset := func(fset string) string {
		return fmt.Sprintf("%-10s", fset)
	}

	if len(modules) != 1 {
		panic("number of modules must be exactly one")
	}

	flag.CommandLine.SetOutput(os.Stdout)
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "%s [subcommand]\n\n", os.Args[0])
		fmt.Fprint(flag.CommandLine.Output(), "subcommands:\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  %s Run using CLI flags.\n", formatFlagset(cli.FLAGSET))
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
	case cli.FLAGSET:
		if err := cli.Parse(subcommandIndex, modules); err != nil {
			return err
		}
		return run(modules)
	case gen.FLAGSET:
		return gen.Generate(subcommandIndex, modules)
	case file.FLAGSET:
		if err := file.Parse(subcommandIndex, modules); err != nil {
			return err
		}
		return run(modules)
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
func run(modules module.Modules) error {
	startLogger.Info("preparing to run the modules")

	// Start each module, exit on error
	for _, m := range modules {
		startLogger.Info("starting", "module", m.Name())
		if err := m.Run(); err != nil {
			startLogger.Error(err, "failed to start module", "module", m.Name())
			panic(err)
		}
	}
	startLogger.Info("all modules started")

	reporter := setupReporter()
	monitor := setupMonitor(reporter, modules)
	var metricServer *http.Server
	if !disableMetricServer {
		metricServer = setupMetricServer(monitor, modules)
	}

	// Start traffic and monitor, with a deadline set to time.Now() + test duration
	background := context.Background()
	deadline := time.Now().Add(duration)
	deadlineCtx, deadlineStop := context.WithDeadline(background, deadline)
	startLogger.Info("traffic will run until", "deadline", deadline)

	if err := monitor.Start(deadlineCtx); err != nil {
		startLogger.Error(err, "failed to start the monitor")
		panic(err)
	}

	if err := traffic.Run(deadlineCtx, modules, reporter); err != nil {
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

	if !disableMetricServer {
		startLogger.Info("stopping metrics server")
		if err := metricServer.Shutdown(stopCtx); err != nil {
			startLogger.Error(err, "metrics server shutdown error")
		}
	}

	startLogger.Info("stopping traffic")
	traffic.AwaitStop()

	startLogger.Info("stopping modules")
	for _, m := range modules {
		if err := m.Stop(); err != nil {
			startLogger.Error(err, "module stop reported an error", "module", m.Name())
		}
	}

	startLogger.Info("finalising report")
	reporter.Finalise()

	return nil
}

func setupReporter() report.Reporter {
	return &report.YAMLReporter{}
}

func setupMonitor(reporter report.Reporter, modules module.Modules) *monitor.Monitor {
	monitor := &monitor.Monitor{Reporter: reporter}

	for _, mod := range modules {
		performancePIDArg := mod.MonitorPerformancePID()
		if *performancePIDArg.Value != module.NO_PERFORMANCE_PID {
			monitor.CPU = cpu.NewLocalCPUMonitor(int32(*performancePIDArg.Value))
			monitor.Memory = memory.NewLocalMemoryMonitor(int32(*performancePIDArg.Value))
		}

		fileArg := mod.MonitorFile()
		if *fileArg.Value != module.NO_LOG_FILE {
			monitor.Log = log.NewLogFileMonitor(*fileArg.Value)
		}

		metricsEndpointArg := mod.MonitorMetricsEndpoint()
		if *metricsEndpointArg.Value != module.NO_METRICS_ENDPOINT {
			// TODO: metric endpoints should be per test module.
			monitor.Metric = metric.NewMetricMonitor(*metricsEndpointArg.Value)
		}
	}

	if monitorDisableMetricTicker {
		monitor.DisableMetricTicker = true
	}

	return monitor
}

func setupMetricServer(monitor *monitor.Monitor, modules module.Modules) *http.Server {
	// Start metric server
	var metricServer *http.Server
	go func() {
		metricServer = &http.Server{
			Addr:    metricAddr,
			Handler: http.DefaultServeMux, // Use the default handler
		}

		// Arbiter metrics
		http.Handle("/metrics", promhttp.Handler())

		// If metric endpoint(s) registered
		for _, mod := range modules {
			if *mod.MonitorMetricsEndpoint().Value != module.NO_METRICS_ENDPOINT {
				http.HandleFunc(fmt.Sprintf("/metrics-%s", mod.Name()), func(w http.ResponseWriter, r *http.Request) {
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

	return metricServer
}
