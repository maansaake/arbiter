package samplemodule

import (
	"time"

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
		args: arg.Args{},
		ops: op.Ops{
			&op.Op{
				Name:     "test",
				Desc:     "Does nothing and returns after a short delay.",
				Disabled: false,
				Rate:     60,
				Do: func() (op.Result, error) {
					time.Sleep(2 * time.Millisecond)
					return op.Result{}, nil
				},
			},
		},
	}
}

func (sm *SampleModule) Name() string {
	return "sample"
}

func (sm *SampleModule) Desc() string {
	return "This is a sample module with a few sample operations."
}

func (lm *SampleModule) MonitorFile() *arg.Arg[string] {
	return &arg.Arg[string]{
		Value: new(string),
	}
}

func (lm *SampleModule) MonitorMetricsEndpoint() *arg.Arg[string] {
	return &arg.Arg[string]{
		Value: new(string),
	}
}

func (lm *SampleModule) MonitorPerformancePID() *arg.Arg[int] {
	return &arg.Arg[int]{
		Value: new(int),
	}
}

func (sm *SampleModule) Args() arg.Args {
	return sm.args
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
