package module

import (
	"tres-bon.se/arbiter/pkg/arg"
	"tres-bon.se/arbiter/pkg/module/op"
)

const (
	NO_PERFORMANCE_PID  = -1
	NO_LOG_FILE         = "none"
	NO_METRICS_ENDPOINT = "none"
)

type Module interface {
	Name() string
	Desc() string
	MonitorPerformancePID() *arg.Arg[int]
	MonitorFile() *arg.Arg[string]
	MonitorMetricsEndpoint() *arg.Arg[string]
	Args() arg.Args
	Ops() op.Ops
	Run() error
	Stop() error
}

type Modules []Module
