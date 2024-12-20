package main

import (
	"os"

	"tres-bon.se/arbiter"
	"tres-bon.se/arbiter/pkg/module"
)

type locksmithModule struct {
	args module.Args
	ops  module.Ops
}

func newLocksmithModule() module.Module {
	return &locksmithModule{
		args: module.Args{
			&module.Arg[bool]{
				Name:  "disable",
				Desc:  "Set if this module should be disabled.",
				Value: new(bool),
			},
			&module.Arg[int]{
				Name:  "routines",
				Desc:  "Sets the number of go-routines.",
				Value: new(int),
			},
		},
		ops: module.Ops{
			&module.Op{
				Name:     "lockunlock",
				Desc:     "Locks and unlocks after a small delay.",
				Disabled: false,
				Rate:     60,
				Do: func() (module.Result, error) {
					return module.Result{}, nil
				},
			},
		},
	}
}

func (lm *locksmithModule) Name() string {
	return "locksmith"
}

func (lm *locksmithModule) Desc() string {
	return "This is a sample module with a few sample operations."
}

func (lm *locksmithModule) MonitorFile() *module.Arg[string] {
	return &module.Arg[string]{
		Value: new(string),
	}
}

func (lm *locksmithModule) MonitorMetricsEndpoint() *module.Arg[string] {
	return &module.Arg[string]{
		Value: new(string),
	}
}

func (lm *locksmithModule) MonitorPerformancePID() *module.Arg[int] {
	return &module.Arg[int]{
		Value: new(int),
	}
}

func (lm *locksmithModule) Args() module.Args {
	return lm.args
}

func (lm *locksmithModule) Ops() module.Ops {
	return lm.ops
}

func (lm *locksmithModule) Run() error {
	return nil
}

func (lm *locksmithModule) Stop() error {
	return nil
}

func main() {
	err := arbiter.Run(module.Modules{newLocksmithModule()})
	if err != nil {
		os.Exit(1)
	}
}
