package samplemod

import (
	"errors"
	"math/rand/v2"
	"time"

	"github.com/maansaake/arbiter/pkg/module"
	"github.com/maansaake/arbiter/pkg/zerologr"
)

const (
	defaultTestDelay time.Duration = 10 * time.Millisecond
	defaultOpRate    uint          = 60
	randRange                      = 100
	brokenOpDelay    time.Duration = 10 * time.Second
)

type SampleModule struct {
	args module.Args
	ops  module.Ops

	testDelay time.Duration
}

func New() module.Module {
	s := &SampleModule{
		testDelay: defaultTestDelay,
	}

	s.args = module.Args{
		&module.Arg[int]{
			Name: "testdelay",
			Desc: "The delay for the 'test' action.",
			Handler: func(v int) {
				s.testDelay = time.Duration(v * int(time.Millisecond))
				zerologr.Info("set value for testdelay", "value", s.testDelay)
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
			Rate: defaultOpRate,
			Do: func() (module.Result, error) {
				time.Sleep(s.testDelay)
				return module.Result{}, nil
			},
		},
		&module.Op{
			Name: "unstable",
			Desc: "Does nothing, sometimes returns an error.",
			Rate: defaultOpRate,
			Do: func() (module.Result, error) {
				//nolint:gosec // just for show
				if rand.IntN(randRange)%2 == 0 {
					return module.Result{}, errors.New("random error")
				}
				return module.Result{}, nil
			},
		},
		&module.Op{
			Name: "broken",
			Desc: "Only returns errors, after 10s.",
			Rate: defaultOpRate,
			Do: func() (module.Result, error) {
				time.Sleep(brokenOpDelay)
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
