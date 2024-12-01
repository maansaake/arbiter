// The cli package implements support for the 'cli' subcommand.
package cli

import (
	"fmt"

	"tres-bon.se/arbiter/pkg/arg"
	"tres-bon.se/arbiter/pkg/module"
	"tres-bon.se/arbiter/pkg/module/op"
	"tres-bon.se/arbiter/pkg/subcommand"
)

const (
	NO_PERFORMANCE_PID = -1
	NO_LOG_FILE        = "none"
	NO_METRIC_ENDPOINT = "none"
)

// Parse command line arguments for the input modules and populate args and
// operations with parsed values.
func Register(subcommandIndex int, modules module.Modules) ([]*subcommand.ModuleMeta, error) {
	moduleMeta := make([]*subcommand.ModuleMeta, len(modules))
	for i, mod := range modules {
		meta := &subcommand.ModuleMeta{}
		meta.PID = NO_PERFORMANCE_PID
		meta.LogFile = NO_LOG_FILE
		meta.MetricEndpoint = NO_METRIC_ENDPOINT

		modArgs := make(arg.Args, 0, len(mod.Args())+(len(mod.Ops())*2))
		modArgs = append(modArgs, mod.Args()...)
		modArgs = append(modArgs, monitorPidArg(mod, meta.PID))
		modArgs = append(modArgs, monitorLogFileArg(mod, meta.LogFile))
		modArgs = append(modArgs, monitorMetricEndpointArg(mod, meta.MetricEndpoint))

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

	return moduleMeta, nil
}

func monitorPidArg(module module.Module, v int) *arg.Arg[int] {
	return &arg.Arg[int]{
		Name:  fmt.Sprintf("%s.monitor.performance.pid", module.Name()),
		Desc:  fmt.Sprintf("A PID to track performance metrics for %s.", module.Name()),
		Value: &v,
	}
}

func monitorLogFileArg(module module.Module, v string) *arg.Arg[string] {
	return &arg.Arg[string]{
		Name:  fmt.Sprintf("%s.monitor.log.file", module.Name()),
		Desc:  fmt.Sprintf("Full path to the log file of %s.", module.Name()),
		Value: &v,
	}
}

func monitorMetricEndpointArg(module module.Module, v string) *arg.Arg[string] {
	return &arg.Arg[string]{
		Name:  fmt.Sprintf("%s.monitor.metric.endpoint", module.Name()),
		Desc:  fmt.Sprintf("Metric endpoint (if any) of %s.", module.Name()),
		Value: &v,
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
