package traffictest

import (
	"tres-bon.se/arbiter/pkg/module"
	"tres-bon.se/arbiter/pkg/module/arg"
	"tres-bon.se/arbiter/pkg/module/op"
)

type testmodule struct {
	name string
	desc string
	args arg.Args
	ops  op.Ops
}

func newTestModule() module.Module {
	return &testmodule{
		name: "testmodule",
		desc: "a module for testing traffic generation",
	}
}

func (tm *testmodule) Name() string {
	return tm.name
}

func (tm *testmodule) Desc() string {
	return tm.desc
}

func (tm *testmodule) Args() arg.Args {
	return tm.args
}

func (tm *testmodule) Ops() op.Ops {
	return tm.ops
}

func (tm *testmodule) Run() error {
	return nil
}

func (tm *testmodule) Stop() error {
	return nil
}
