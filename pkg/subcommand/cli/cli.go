// Package cli implements support for the 'cli' subcommand.
package cli

import (
	"fmt"
	"strings"

	"github.com/maansaake/arbiter/pkg/module"
)

const argsPerOp = 2 // each op contributes a disable flag and a rate flag

// Parse command line arguments for the input modules.
func Parse(args []string, modules module.Modules) (module.Metadata, error) {
	metadata := make(module.Metadata, len(modules))
	for i, mod := range modules {
		meta := &module.Meta{Module: mod}

		modArgs := make(module.Args, 0, len(mod.Args())+len(mod.Ops())*argsPerOp)
		modArgs = append(modArgs, mod.Args()...)

		// Add operation args
		for _, op := range mod.Ops() {
			modArgs = append(modArgs, disableArg(op))
			modArgs = append(modArgs, rateArg(op))
		}

		if err := Register(strings.ToLower(mod.Name()), modArgs); err != nil {
			return nil, err
		}

		metadata[i] = meta
	}

	return metadata, ParseArgs(args)
}

func disableArg(op *module.Op) *module.Arg[bool] {
	return &module.Arg[bool]{
		Name:  fmt.Sprintf("op.%s.disable", strings.ToLower(op.Name)),
		Desc:  fmt.Sprintf("Disable %s.", op.Name),
		Value: &op.Disabled,
	}
}

func rateArg(op *module.Op) *module.Arg[uint] {
	return &module.Arg[uint]{
		Name:  fmt.Sprintf("op.%s.rate", strings.ToLower(op.Name)),
		Desc:  fmt.Sprintf("Rate of %s per minute.", op.Name),
		Value: &op.Rate,
	}
}
