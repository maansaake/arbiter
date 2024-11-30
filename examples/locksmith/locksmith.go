package main

import (
	"os"

	"tres-bon.se/arbiter"
	"tres-bon.se/arbiter/pkg/arg"
	"tres-bon.se/arbiter/pkg/module"
	"tres-bon.se/arbiter/pkg/module/op"
)

type locksmithModule struct {
	args arg.Args
	ops  op.Ops
}

func newLocksmithModule() module.Module {
	return &locksmithModule{
		args: arg.Args{
			&arg.Arg[bool]{
				Name:  "disable",
				Desc:  "Set if this module should be disabled.",
				Value: new(bool),
			},
			&arg.Arg[int]{
				Name:  "routines",
				Desc:  "Sets the number of go-routines.",
				Value: new(int),
			},
		},
		ops: op.Ops{
			&op.Op{
				Name:     "lockunlock",
				Desc:     "Locks and unlocks after a small delay.",
				Disabled: false,
				Rate:     60,
				Do: func() (op.Result, error) {
					return op.Result{}, nil
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

func (lm *locksmithModule) MonitorFile() *arg.Arg[string] {
	return &arg.Arg[string]{
		Value: new(string),
	}
}

func (lm *locksmithModule) MonitorMetricsEndpoint() *arg.Arg[string] {
	return &arg.Arg[string]{
		Value: new(string),
	}
}

func (lm *locksmithModule) MonitorPerformancePID() *arg.Arg[int] {
	return &arg.Arg[int]{
		Value: new(int),
	}
}

func (lm *locksmithModule) Args() arg.Args {
	return lm.args
}

func (lm *locksmithModule) Ops() op.Ops {
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
