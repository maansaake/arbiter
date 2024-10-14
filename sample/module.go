package samplemodule

import (
	"tres-bon.se/arbiter/pkg/arg"
	"tres-bon.se/arbiter/pkg/module"
	"tres-bon.se/arbiter/pkg/module/op"
)

type SampleModule struct {
	args arg.Args
	ops  op.Ops
}

func NewSampleModule() module.Module {
	return &SampleModule{
		args: arg.Args{
			&arg.Arg[bool]{
				Name: "enabled",
				Desc: "Sets if this module should be enabled.",
			},
			&arg.Arg[int]{
				Name: "routines",
				Desc: "Sets the number of go-routines",
			},
		},
		ops: op.Ops{
			&op.Op{
				Name: "lockunlock",
				Desc: "Locks and unlocks after a small delay",
				Do: func() (op.Result, error) {
					return op.Result{}, nil
				},
			},
		},
	}
}

func (sm *SampleModule) Name() string {
	return "sample"
}

func (sm *SampleModule) Args() arg.Args {
	return sm.args
}

func (sm *SampleModule) Desc() string {
	return "This is a sample module with a few sample operations."
}

func (sm *SampleModule) Ops() op.Ops {
	return sm.ops
}

func (sm *SampleModule) Run() error {
	return nil
}

func (sm *SampleModule) Stop() error {
	return nil
}
