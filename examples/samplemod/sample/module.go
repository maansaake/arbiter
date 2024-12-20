package samplemodule

import (
	"errors"
	"math/rand"
	"time"

	"tres-bon.se/arbiter/pkg/module"
	"tres-bon.se/arbiter/pkg/module/arg"
	"tres-bon.se/arbiter/pkg/module/op"
	"tres-bon.se/arbiter/pkg/zerologr"
)

type SampleModule struct {
	args arg.Args
	ops  op.Ops

	testDelayMs time.Duration
}

func NewSampleModule() module.Module {
	s := &SampleModule{
		testDelayMs: 10 * time.Millisecond,
	}

	s.args = arg.Args{
		&arg.Arg[int]{
			Name: "testdelay",
			Desc: "The delay for the 'test' action.",
			Handler: func(v int) {
				s.testDelayMs = time.Duration(v * int(time.Millisecond))
				zerologr.Info("set value for testdelay", "value", s.testDelayMs)
			},
		},
	}

	s.ops = op.Ops{
		&op.Op{
			Name: "test",
			Desc: "Does nothing and returns after a configurable delay.",
			Rate: 60,
			Do: func() (op.Result, error) {
				time.Sleep(s.testDelayMs)
				return op.Result{}, nil
			},
		},
		&op.Op{
			Name: "unstable",
			Desc: "Does nothing, sometimes returns an error.",
			Rate: 60,
			Do: func() (op.Result, error) {
				//nolint:gosec // just for show
				if rand.Intn(100)%2 == 0 {
					return op.Result{}, errors.New("random error")
				}
				return op.Result{}, nil
			},
		},
		&op.Op{
			Name: "broken",
			Desc: "Only returns errors, after 10s.",
			Rate: 60,
			Do: func() (op.Result, error) {
				time.Sleep(10 * time.Second)
				return op.Result{}, errors.New("permanent failure")
			},
		},
	}

	return s
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

// INTERNAL
