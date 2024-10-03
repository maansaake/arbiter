package arbiter

import (
	"tres-bon.se/arbiter/pkg/arg"
	"tres-bon.se/arbiter/pkg/module"
)

func Run(modules module.Modules) error {
	args := make(arg.Args, 0, 0)
	for _, m := range modules {
		args = append(args, m.Args()...)
	}

	if err := arg.ParseArgs(args); err != nil {

	}

	return nil
}
