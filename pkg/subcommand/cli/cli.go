// The cli package implements support for the 'cli' subcommand.
package cli

import (
	"fmt"
	"os"
	"strings"

	"tres-bon.se/arbiter/pkg/module"
	"tres-bon.se/arbiter/pkg/monitor"
	"tres-bon.se/arbiter/pkg/monitor/cpu"
	"tres-bon.se/arbiter/pkg/monitor/trigger"
	"tres-bon.se/arbiter/pkg/subcommand"
)

// Parse command line arguments for the input modules.
func Parse(subcommandIndex int, modules module.Modules) (subcommand.Metadata, error) {
	metadata := make(subcommand.Metadata, len(modules))
	for i, mod := range modules {
		meta := &subcommand.Meta{Module: mod, MonitorOpt: monitor.DefaultOpt()}

		numMonitorArgs := 8
		monitorArgs := make(module.Args, numMonitorArgs)
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

		modArgs := make(module.Args, 0, len(mod.Args())+(len(mod.Ops())*2)+len(monitorArgs))
		modArgs = append(modArgs, mod.Args()...)
		modArgs = append(modArgs, monitorArgs...)

		// Add operation args
		for _, op := range mod.Ops() {
			modArgs = append(modArgs, disableArg(op))
			modArgs = append(modArgs, rateArg(op))
		}

		if err := Register(strings.ToLower(mod.Name()), modArgs); err != nil {
			return nil, err
		}

		metadata[i] = meta
	}

	return metadata, ParseArgs(os.Args[subcommandIndex+1:])
}

func monitorPidArg(v int) *module.Arg[int] {
	return &module.Arg[int]{
		Name:  "monitor.performance.pid",
		Desc:  "A PID to track performance metrics.",
		Value: &v,
		Valid: cpu.ValidPID,
	}
}

func cpuTrigger(monitorOpt *monitor.Opt) *module.Arg[string] {
	// Define an argument that can be set one or more times, each adding to the input monitorOpt.
	return &module.Arg[string]{
		Name:    "monitor.cpu.trigger",
		Desc:    "Trigger(s) for CPU levels.",
		Valid:   trigger.ValidCPUTrigger,
		Handler: monitorOpt.CPUTriggerFromCmdline,
	}
}

func vmsTrigger(monitorOpt *monitor.Opt) *module.Arg[string] {
	return &module.Arg[string]{
		Name:    "monitor.vms.trigger",
		Desc:    "Trigger(s) for VMS levels.",
		Valid:   trigger.ValidVMSTrigger,
		Handler: monitorOpt.VMSTriggerFromCmdline,
	}
}

func rssTrigger(monitorOpt *monitor.Opt) *module.Arg[string] {
	return &module.Arg[string]{
		Name:    "monitor.rss.trigger",
		Desc:    "Trigger(s) for RSS levels.",
		Valid:   trigger.ValidRSSTrigger,
		Handler: monitorOpt.RSSTriggerFromCmdline,
	}
}

func monitorLogFileArg(v string) *module.Arg[string] {
	return &module.Arg[string]{
		Name:  "monitor.log.file",
		Desc:  "Full path to a log file.",
		Value: &v,
	}
}

func logFileTrigger(monitorOpt *monitor.Opt) *module.Arg[string] {
	return &module.Arg[string]{
		Name:    "monitor.log.trigger",
		Desc:    "Trigger(s) for log files.",
		Valid:   trigger.ValidLogFileTrigger,
		Handler: monitorOpt.LogFileTriggerFromCmdline,
	}
}

func monitorMetricEndpointArg(v string) *module.Arg[string] {
	return &module.Arg[string]{
		Name:  "monitor.metric.endpoint",
		Desc:  "Metric endpoint (if any).",
		Value: &v,
	}
}

func metricTrigger(monitorOpt *monitor.Opt) *module.Arg[string] {
	return &module.Arg[string]{
		Name:    "monitor.metric.trigger",
		Desc:    "Trigger(s) for metrics.",
		Valid:   trigger.ValidMetricTrigger,
		Handler: monitorOpt.MetricTriggerFromCmdline,
	}
}

func disableArg(op *module.Op) *module.Arg[bool] {
	return &module.Arg[bool]{
		Name:  fmt.Sprintf("op.%s.disable", strings.ToLower(op.Name)),
		Desc:  fmt.Sprintf("Disable %s.", op.Name),
		Value: &op.Disabled,
	}
}

func rateArg(op *module.Op) *module.Arg[uint] {
	return &module.Arg[uint]{
		Name:  fmt.Sprintf("op.%s.rate", strings.ToLower(op.Name)),
		Desc:  fmt.Sprintf("Rate of %s per minute.", op.Name),
		Value: &op.Rate,
	}
}
