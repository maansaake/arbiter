package testmodule

import (
	"tres-bon.se/arbiter/pkg/module"
	"tres-bon.se/arbiter/pkg/module/arg"
	"tres-bon.se/arbiter/pkg/module/op"
)

type TestModule struct {
	SetName string
	SetDesc string
	SetArgs arg.Args
	SetOps  op.Ops
}

func NewTestModule() module.Module {
	return &TestModule{
		SetName: "testmodule",
		SetDesc: "a module for testing traffic generation",
	}
}

func (tm *TestModule) Name() string {
	return tm.SetName
}

func (tm *TestModule) Desc() string {
	return tm.SetDesc
}

func (tm *TestModule) MonitorFile() *arg.Arg[string] {
	s := new(string)
	*s = module.NO_LOG_FILE
	return &arg.Arg[string]{
		Value: s,
	}
}

func (tm *TestModule) MonitorMetricsEndpoint() *arg.Arg[string] {
	s := new(string)
	*s = module.NO_METRICS_ENDPOINT
	return &arg.Arg[string]{
		Value: s,
	}
}

func (tm *TestModule) MonitorPerformancePID() *arg.Arg[int] {
	i := new(int)
	*i = module.NO_PERFORMANCE_PID
	return &arg.Arg[int]{
		Value: i,
	}
}

func (tm *TestModule) Args() arg.Args {
	return tm.SetArgs
}

func (tm *TestModule) Ops() op.Ops {
	return tm.SetOps
}

func (tm *TestModule) Run() error {
	return nil
}

func (tm *TestModule) Stop() error {
	return nil
}
