// The cli package implements support for the 'cli' subcommand.
package cli

import (
	"fmt"
	"os"

	"tres-bon.se/arbiter/pkg/arg"
	"tres-bon.se/arbiter/pkg/module"
	"tres-bon.se/arbiter/pkg/module/op"
	"tres-bon.se/arbiter/pkg/monitor"
	"tres-bon.se/arbiter/pkg/monitor/trigger"
	"tres-bon.se/arbiter/pkg/subcommand"
)

// Register command line arguments for the input modules.
func Register(subcommandIndex int, modules module.Modules) ([]*subcommand.Meta, error) {
	moduleMeta := make([]*subcommand.Meta, len(modules))
	for i, mod := range modules {
		meta := &subcommand.Meta{Module: mod, MonitorOpt: &monitor.Opt{}}
		meta.MonitorOpt.PID = monitor.NO_PERFORMANCE_PID
		meta.MonitorOpt.LogFile = monitor.NO_LOG_FILE
		meta.MonitorOpt.MetricEndpoint = monitor.NO_METRIC_ENDPOINT

		numMonitorArgs := 8
		monitorArgs := make(arg.Args, numMonitorArgs)
		// Performance
		monitorArgs[0] = monitorPidArg(meta.MonitorOpt.PID)
		monitorArgs[1] = cpuTrigger(meta.MonitorOpt)
		monitorArgs[2] = vmsTrigger(meta.MonitorOpt)
		monitorArgs[3] = rssTrigger(meta.MonitorOpt)

		// Logs
		monitorArgs[4] = monitorLogFileArg(meta.MonitorOpt.LogFile)
		monitorArgs[5] = logFileTrigger(meta.MonitorOpt)

		// Metrics
		monitorArgs[6] = monitorMetricEndpointArg(meta.MonitorOpt.MetricEndpoint)
		monitorArgs[7] = metricTrigger(meta.MonitorOpt)

		modArgs := make(arg.Args, 0, len(mod.Args())+(len(mod.Ops())*2)+len(monitorArgs))
		modArgs = append(modArgs, mod.Args()...)
		modArgs = append(modArgs, monitorArgs...)

		// Add operation args
		for _, op := range mod.Ops() {
			modArgs = append(modArgs, disableArg(op))
			modArgs = append(modArgs, rateArg(op))
		}

		if err := arg.Register(mod.Name(), modArgs); err != nil {
			return nil, err
		}

		moduleMeta[i] = meta
	}

	return moduleMeta, arg.Parse(os.Args[subcommandIndex+1:])
}

func monitorPidArg(v int) *arg.Arg[int] {
	return &arg.Arg[int]{
		Name:  "monitor.performance.pid",
		Desc:  "A PID to track performance metrics.",
		Value: &v,
	}
}

func cpuTrigger(monitorOpt *monitor.Opt) *arg.Arg[string] {
	// Define an argument that can be set one or more times, each adding to the input monitorOpt.
	return &arg.Arg[string]{
		Name:    "monitor.cpu.trigger",
		Desc:    "Trigger(s) for CPU levels.",
		Valid:   trigger.ValidCPUTrigger,
		Handler: monitorOpt.CPUTriggerFromCmdline,
	}
}

func vmsTrigger(monitorOpt *monitor.Opt) *arg.Arg[string] {
	return &arg.Arg[string]{
		Name:    "monitor.vms.trigger",
		Desc:    "Trigger(s) for VMS levels.",
		Valid:   trigger.ValidVMSTrigger,
		Handler: monitorOpt.VMSTriggerFromCmdline,
	}
}

func rssTrigger(monitorOpt *monitor.Opt) *arg.Arg[string] {
	return &arg.Arg[string]{
		Name:    "monitor.rss.trigger",
		Desc:    "Trigger(s) for RSS levels.",
		Valid:   trigger.ValidRSSTrigger,
		Handler: monitorOpt.RSSTriggerFromCmdline,
	}
}

func monitorLogFileArg(v string) *arg.Arg[string] {
	return &arg.Arg[string]{
		Name:  "monitor.log.file",
		Desc:  "Full path to a log file.",
		Value: &v,
	}
}

func logFileTrigger(monitorOpt *monitor.Opt) *arg.Arg[string] {
	return &arg.Arg[string]{
		Name:    "monitor.log.trigger",
		Desc:    "Trigger(s) for log files.",
		Valid:   trigger.ValidLogFileTrigger,
		Handler: monitorOpt.LogFileTriggerFromCmdline,
	}
}

func monitorMetricEndpointArg(v string) *arg.Arg[string] {
	return &arg.Arg[string]{
		Name:  "monitor.metric.endpoint",
		Desc:  "Metric endpoint (if any).",
		Value: &v,
	}
}

func metricTrigger(monitorOpt *monitor.Opt) *arg.Arg[string] {
	return &arg.Arg[string]{
		Name:    "monitor.metric.trigger",
		Desc:    "Trigger(s) for metrics.",
		Valid:   trigger.ValidMetricTrigger,
		Handler: monitorOpt.MetricTriggerFromCmdline,
	}
}

func disableArg(op *op.Op) *arg.Arg[bool] {
	return &arg.Arg[bool]{
		Name:  fmt.Sprintf("op.%s.disable", op.Name),
		Desc:  fmt.Sprintf("Disable %s.", op.Name),
		Value: &op.Disabled,
	}
}

func rateArg(op *op.Op) *arg.Arg[uint] {
	return &arg.Arg[uint]{
		Name:  fmt.Sprintf("op.%s.rate", op.Name),
		Desc:  fmt.Sprintf("Rate of %s per minute.", op.Name),
		Value: &op.Rate,
	}
}
