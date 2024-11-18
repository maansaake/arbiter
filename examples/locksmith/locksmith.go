package main

import (
	"os"

	"tres-bon.se/arbiter"
	"tres-bon.se/arbiter/pkg/module"
	"tres-bon.se/arbiter/pkg/module/arg"
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
				Desc:  "Sets the number of go-routines",
				Value: new(int),
			},
		},
		ops: op.Ops{
			&op.Op{
				Name:     "lockunlock",
				Desc:     "Locks and unlocks after a small delay",
				Disabled: false,
				Rate:     60,
				Do: func() (op.Result, error) {
					return op.Result{}, nil
				},
			},
		},
	}
}

func (sm *locksmithModule) Name() string {
	return "locksmith"
}

func (sm *locksmithModule) Args() arg.Args {
	return sm.args
}

func (sm *locksmithModule) Desc() string {
	return "This is a sample module with a few sample operations."
}

func (sm *locksmithModule) Ops() op.Ops {
	return sm.ops
}

func (sm *locksmithModule) Run() error {
	return nil
}

func (sm *locksmithModule) Stop() error {
	return nil
}

func main() {
	err := arbiter.Run(module.Modules{newLocksmithModule()})
	if err != nil {
		os.Exit(1)
	}
}
