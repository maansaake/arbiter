// Arbiter runs system tests on the target system.
//
// By providing custom modules, Arbiter is able to generate traffic and monitor
// any system, measuing CPU, memory, metrics, and logs. Arbiter can judge a
// system based on those four parameters. Add rates for operations and
// thresholds to verify the software is staying within expected boundaries.
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
	modulearg "tres-bon.se/arbiter/pkg/module/arg"
	"tres-bon.se/arbiter/pkg/monitor"
	"tres-bon.se/arbiter/pkg/monitor/cpu"
	"tres-bon.se/arbiter/pkg/monitor/log"
	"tres-bon.se/arbiter/pkg/monitor/memory"
	"tres-bon.se/arbiter/pkg/monitor/metric"
	"tres-bon.se/arbiter/pkg/report"
	"tres-bon.se/arbiter/pkg/traffic"
	"tres-bon.se/arbiter/pkg/zerologr"
)

const (
	FLAGSET_CLI  = modulearg.FLAGSET
	FLAGSET_GEN  = "generate"
	FLAGSET_FILE = "file"

	MONITOR_PID_DEFAULT              = -1
	MONITOR_FILE_DEFAULT             = "none"
	MONITOR_METRICS_ENDPOINT_DEFAULT = "none"
)

var (
	// global flag vars.
	duration    time.Duration = time.Minute * 5
	monitorPid  int           = MONITOR_PID_DEFAULT
	monitorFile string        = MONITOR_FILE_DEFAULT
	// TODO: metric endpoints should be per test module.
	monitorMetricsEndpoint string = MONITOR_METRICS_ENDPOINT_DEFAULT

	// subcommand parsing vars.
	subcommands     = []string{FLAGSET_CLI, FLAGSET_GEN, FLAGSET_FILE}
	subcommandIndex = -1

	// log.
	startLogger = zerologr.New(&zerologr.Opts{Console: true, V: 10}).WithName("start")

	// metrics.
	metricAddr = ":8888"
	// metricPrefix = "arbiter"
)

// Runs the Arbiter. Blocks until SIGINT, SIGTERM or when the test duration
// runs out (5 minute default).
func Run(modules module.Modules) error {
	if len(modules) != 1 {
		panic("number of modules must be exactly one")
	}

	flag.CommandLine.SetOutput(os.Stdout)
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "%s [subcommand]\n\n", os.Args[0])
		fmt.Fprint(flag.CommandLine.Output(), "subcommands:\n")
		fmt.Fprint(flag.CommandLine.Output(), "  cli      Run using CLI flags.\n")
		fmt.Fprint(flag.CommandLine.Output(), "  generate Generate a test model file.\n")
		fmt.Fprint(flag.CommandLine.Output(), "  file     Run from a test model file.\n")
		fmt.Fprint(flag.CommandLine.Output(), "\n")
		fmt.Fprint(flag.CommandLine.Output(), "global flags:\n")
		flag.PrintDefaults()
	}

	// Global flags
	flag.DurationVar(&duration, "duration", duration, "The duration of the test run, minimum 30 seconds.")
	flag.IntVar(&monitorPid, "monitor.performance.pid", monitorPid, "A PID to monitor resource usage (CPU & memory) of during the test run.")
	flag.StringVar(&monitorFile, "monitor.log.file", monitorFile, "A file to stream log entries from.")
	flag.StringVar(&monitorMetricsEndpoint, "monitor.metric.endpoint", monitorMetricsEndpoint, "An endpoint to fetch metrics from.")

	// To trigger on --help and parse global flags
	flag.Parse()

	// Verify a subcommand has been invoked.
	subcommandIndex = slices.IndexFunc(os.Args, func(e string) bool {
		return slices.Contains(subcommands, e)
	})
	parseError := false
	if subcommandIndex == -1 {
		parseError = true
		fmt.Fprint(flag.CommandLine.Output(), "no subcommand given\n")
	}

	if duration < 30*time.Second {
		parseError = true
		fmt.Fprint(flag.CommandLine.Output(), "test duration was less than 30 seconds\n")
	}

	if parseError {
		flag.CommandLine.SetOutput(os.Stderr)
		flag.Usage()
		os.Exit(1)
	}

	// Check invoked subcommand
	switch os.Args[subcommandIndex] {
	case modulearg.FLAGSET:
		return handleCli(modules)
	case FLAGSET_GEN:
		return handleGen(modules)
	case FLAGSET_FILE:
		return handleFile(modules)
	default:
		flag.CommandLine.SetOutput(os.Stderr)
		fmt.Fprintf(flag.CommandLine.Output(), "subcommand not found: %s\n", os.Args[1])
		flag.Usage()
		os.Exit(1)
	}

	return nil
}

// Handle the CLI subcommand call, registering cmd line flags for both module
// arguments and their operations and parsing them. Runs the modules with the
// input flags.
func handleCli(modules module.Modules) error {
	for _, m := range modules {
		for _, a := range m.Args() {
			modulearg.Register(m.Name(), a)
		}

		// Add operation args
		for _, op := range m.Ops() {
			modulearg.RegisterOp(m.Name(), op)
		}
	}

	modulearg.Parse(os.Args[subcommandIndex+1:])

	startLogger.Info("parsed CLI arguments")

	return run(modules)
}

// Handles the generate subcommand. Generates a test model file based on the
// input modules.
func handleGen(_ module.Modules) error {
	var output string
	fs := flag.NewFlagSet(FLAGSET_GEN, flag.ExitOnError)
	fs.StringVar(&output, "output", "./arbiter.yaml", "Output path for the generated test model file.")
	err := fs.Parse(os.Args[subcommandIndex+1:])
	if err != nil {
		fs.SetOutput(os.Stderr)
		fs.Usage()
		os.Exit(1)
	}

	// TODO: generate using input modules
	panic("not implemented")
}

// Handles the file subcommand, parsing the input test model file and running
// the modules with the file's settings.
func handleFile(_ module.Modules) error {
	outputPath := "./arbiter.yaml"
	fs := flag.NewFlagSet(FLAGSET_FILE, flag.ExitOnError)
	fs.StringVar(&outputPath, "path", outputPath, "Path to a test model file.")
	err := fs.Parse(os.Args[subcommandIndex+1:])
	if err != nil {
		fs.SetOutput(os.Stderr)
		fs.Usage()
		os.Exit(1)
	}

	// TODO: parse and run from file
	panic("not implemented")
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

	// Hook up monitor and reporter
	// TODO: add choice of reporter implementation - future, new arbiter.Run(...)
	reporter := &report.YAMLReporter{}
	// TODO: add threshold support
	// TODO: add choice of monitor implementations - future, new arbiter.Run(...)
	monitor := &monitor.Monitor{Reporter: reporter}

	if monitorPid != MONITOR_PID_DEFAULT {
		monitor.CPU = cpu.NewLocalCPUMonitor(int32(monitorPid))
		monitor.Memory = memory.NewLocalMemoryMonitor(int32(monitorPid))
	}

	if monitorFile != MONITOR_FILE_DEFAULT {
		monitor.Log = log.NewLogFileMonitor(monitorFile)
	}

	// TODO: metric endpoints should be per test module.
	if monitorMetricsEndpoint != MONITOR_METRICS_ENDPOINT_DEFAULT {
		monitor.Metric = metric.NewMetricMonitor(monitorMetricsEndpoint)
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

	// Start metric server
	var metricServer *http.Server
	go func() {
		metricServer = &http.Server{
			Addr:    metricAddr,
			Handler: http.DefaultServeMux, // Use the default handler
		}
		http.Handle("/metrics", promhttp.Handler())
		for _, module := range modules {
			http.HandleFunc(fmt.Sprintf("/metrics-%s", module.Name()), func(w http.ResponseWriter, r *http.Request) {
				_, err := w.Write(monitor.LatestRawMetrics(module.Name()))
				if err != nil {
					w.WriteHeader(500)
				}
			})
		}

		startLogger.Info("running metrics server on", "address", metricAddr)
		if err := metricServer.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			startLogger.Error(err, "unexpected error")
		} else {
			startLogger.Info("metrics server shut down")
		}
	}()

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

	startLogger.Info("stopping metrics server")
	if err := metricServer.Shutdown(stopCtx); err != nil {
		startLogger.Error(err, "metrics server shutdown error")
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
