// The cli package implements support for the 'cli' subcommand.
package cli

import (
	"fmt"

	"tres-bon.se/arbiter/pkg/arg"
	"tres-bon.se/arbiter/pkg/module"
	"tres-bon.se/arbiter/pkg/module/op"
)

// Parse command line arguments for the input modules and populate args and
// operations with parsed values.
func Register(subcommandIndex int, modules module.Modules) error {
	for _, mod := range modules {
		modArgs := make(arg.Args, 0, len(mod.Args())+(len(mod.Ops())*2))
		modArgs = append(modArgs, mod.Args()...)

		// Add operation args
		for _, op := range mod.Ops() {
			modArgs = append(modArgs, disableArg(op))
			modArgs = append(modArgs, rateArg(op))
		}

		if err := arg.Register(mod.Name(), modArgs); err != nil {
			return err
		}
	}

	return nil
}

func disableArg(op *op.Op) *arg.Arg[bool] {
	return &arg.Arg[bool]{
		Name:  disableName(op),
		Desc:  disableDesc(op),
		Value: &op.Disabled,
	}
}

func rateArg(op *op.Op) *arg.Arg[uint] {
	return &arg.Arg[uint]{
		Name:  rateName(op),
		Desc:  rateDesc(op),
		Value: &op.Rate,
	}
}

func disableName(op *op.Op) string {
	return fmt.Sprintf("%s.disable", op.Name)
}

func disableDesc(op *op.Op) string {
	return fmt.Sprintf("Disable %s.", op.Name)
}

func rateName(op *op.Op) string {
	return fmt.Sprintf("%s.rate", op.Name)
}

func rateDesc(op *op.Op) string {
	return fmt.Sprintf("Rate of %s per minute.", op.Name)
}
