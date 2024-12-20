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
