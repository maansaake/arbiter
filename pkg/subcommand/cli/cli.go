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

const (
	NO_PERFORMANCE_PID = -1
	NO_LOG_FILE        = "none"
	NO_METRIC_ENDPOINT = "none"
)

// Register command line arguments for the input modules.
func Register(subcommandIndex int, modules module.Modules) ([]*subcommand.ModuleMeta, error) {
	moduleMeta := make([]*subcommand.ModuleMeta, len(modules))
	for i, mod := range modules {
		meta := &subcommand.ModuleMeta{Module: mod, ModuleInfo: &monitor.ModuleInfo{}}
		meta.PID = NO_PERFORMANCE_PID
		meta.LogFile = NO_LOG_FILE
		meta.MetricEndpoint = NO_METRIC_ENDPOINT

		modArgs := make(arg.Args, 0, len(mod.Args())+(len(mod.Ops())*2))
		modArgs = append(modArgs, mod.Args()...)

		// Performance
		modArgs = append(modArgs, monitorPidArg(mod, meta.PID))
		modArgs = append(modArgs, cpuTrigger(mod, meta.ModuleInfo))
		modArgs = append(modArgs, vmsTrigger(mod, meta.ModuleInfo))
		modArgs = append(modArgs, rssTrigger(mod, meta.ModuleInfo))

		// Logs
		modArgs = append(modArgs, monitorLogFileArg(mod, meta.LogFile))
		modArgs = append(modArgs, logFileTrigger(mod, meta.ModuleInfo))

		// Metrics
		modArgs = append(modArgs, monitorMetricEndpointArg(mod, meta.MetricEndpoint))
		modArgs = append(modArgs, metricTrigger(mod, meta.ModuleInfo))

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

func monitorPidArg(module module.Module, v int) *arg.Arg[int] {
	return &arg.Arg[int]{
		Name:  "monitor.performance.pid",
		Desc:  fmt.Sprintf("A PID to track performance metrics for %s.", module.Name()),
		Value: &v,
	}
}

func cpuTrigger(module module.Module, moduleInfo *monitor.ModuleInfo) *arg.Arg[string] {
	// Define an argument that can be set one or more times, each adding to the input moduleInfo.
	return &arg.Arg[string]{
		Name:    "monitor.cpu.trigger",
		Desc:    "Trigger(s) for CPU levels.",
		Valid:   trigger.ValidCPUTrigger,
		Handler: moduleInfo.RegisterCPUTrigger,
	}
}

func vmsTrigger(module module.Module, moduleInfo *monitor.ModuleInfo) *arg.Arg[string] {
	return &arg.Arg[string]{
		Name:    "monitor.vms.trigger",
		Desc:    "Trigger(s) for VMS levels.",
		Valid:   trigger.ValidVMSTrigger,
		Handler: moduleInfo.RegisterVMSTrigger,
	}
}

func rssTrigger(module module.Module, moduleInfo *monitor.ModuleInfo) *arg.Arg[string] {
	return &arg.Arg[string]{
		Name:    "monitor.cpu.trigger",
		Desc:    "Trigger(s) for CPU levels.",
		Valid:   trigger.ValidRSSTrigger,
		Handler: moduleInfo.RegisterRSSTrigger,
	}
}

func monitorLogFileArg(module module.Module, v string) *arg.Arg[string] {
	return &arg.Arg[string]{
		Name:  "monitor.log.file",
		Desc:  fmt.Sprintf("Full path to the log file of %s.", module.Name()),
		Value: &v,
	}
}

func logFileTrigger(module module.Module, moduleInfo *monitor.ModuleInfo) *arg.Arg[string] {
	return &arg.Arg[string]{
		Name:    "monitor.cpu.trigger",
		Desc:    "Trigger(s) for CPU levels.",
		Valid:   trigger.ValidLogFileTrigger,
		Handler: moduleInfo.RegisterLogFileTrigger,
	}
}

func monitorMetricEndpointArg(module module.Module, v string) *arg.Arg[string] {
	return &arg.Arg[string]{
		Name:  "monitor.metric.endpoint",
		Desc:  fmt.Sprintf("Metric endpoint (if any) of %s.", module.Name()),
		Value: &v,
	}
}

func metricTrigger(module module.Module, moduleInfo *monitor.ModuleInfo) *arg.Arg[string] {
	return &arg.Arg[string]{
		Name:    "monitor.cpu.trigger",
		Desc:    "Trigger(s) for CPU levels.",
		Valid:   trigger.ValidMetricTrigger,
		Handler: moduleInfo.RegisterMetricTrigger,
	}
}

func disableArg(op *op.Op) *arg.Arg[bool] {
	return &arg.Arg[bool]{
		Name:  fmt.Sprintf("%s.disable", op.Name),
		Desc:  fmt.Sprintf("Disable %s.", op.Name),
		Value: &op.Disabled,
	}
}

func rateArg(op *op.Op) *arg.Arg[uint] {
	return &arg.Arg[uint]{
		Name:  fmt.Sprintf("%s.rate", op.Name),
		Desc:  fmt.Sprintf("Rate of %s per minute.", op.Name),
		Value: &op.Rate,
	}
}
