package samplemodule

import (
	"errors"
	"math/rand"
	"time"

	"tres-bon.se/arbiter/pkg/module"
	"tres-bon.se/arbiter/pkg/zerologr"
)

type SampleModule struct {
	args module.Args
	ops  module.Ops

	testDelayMs time.Duration
}

func NewSampleModule() module.Module {
	s := &SampleModule{
		testDelayMs: 10 * time.Millisecond,
	}

	s.args = module.Args{
		&module.Arg[int]{
			Name: "testdelay",
			Desc: "The delay for the 'test' action.",
			Handler: func(v int) {
				s.testDelayMs = time.Duration(v * int(time.Millisecond))
				zerologr.Info("set value for testdelay", "value", s.testDelayMs)
			},
		},
		&module.Arg[int]{
			Name:     "important",
			Desc:     "A very important argument.",
			Required: true,
			Value:    new(int),
		},
	}

	s.ops = module.Ops{
		&module.Op{
			Name: "test",
			Desc: "Does nothing and returns after a configurable delay.",
			Rate: 60,
			Do: func() (module.Result, error) {
				time.Sleep(s.testDelayMs)
				return module.Result{}, nil
			},
		},
		&module.Op{
			Name: "unstable",
			Desc: "Does nothing, sometimes returns an error.",
			Rate: 60,
			Do: func() (module.Result, error) {
				//nolint:gosec // just for show
				if rand.Intn(100)%2 == 0 {
					return module.Result{}, errors.New("random error")
				}
				return module.Result{}, nil
			},
		},
		&module.Op{
			Name: "broken",
			Desc: "Only returns errors, after 10s.",
			Rate: 60,
			Do: func() (module.Result, error) {
				time.Sleep(10 * time.Second)
				return module.Result{}, errors.New("permanent failure")
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

func (lm *SampleModule) MonitorFile() *module.Arg[string] {
	return &module.Arg[string]{
		Value: new(string),
	}
}

func (lm *SampleModule) MonitorMetricsEndpoint() *module.Arg[string] {
	return &module.Arg[string]{
		Value: new(string),
	}
}

func (lm *SampleModule) MonitorPerformancePID() *module.Arg[int] {
	return &module.Arg[int]{
		Value: new(int),
	}
}

func (sm *SampleModule) Args() module.Args {
	return sm.args
}

func (sm *SampleModule) Ops() module.Ops {
	return sm.ops
}

func (sm *SampleModule) Run() error {
	return nil
}

func (sm *SampleModule) Stop() error {
	return nil
}

// INTERNAL
