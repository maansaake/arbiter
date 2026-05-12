// Package cli implements support for the 'cli' subcommand.
package cli

import (
	"fmt"
	"strings"

	"github.com/maansaake/arbiter/pkg/module"
	"github.com/spf13/cobra"
)

const argsPerOp = 2 // each op contributes a disable flag and a rate flag

// NewCommand creates a cobra command for the 'cli' subcommand populated with
// flags derived from the given modules. The provided run function is called
// with the resolved metadata when the command executes.
func NewCommand(modules module.Modules, run func(module.Metadata) error) (*cobra.Command, error) {
	metadata := make(module.Metadata, len(modules))

	var required []string

	cmd := &cobra.Command{
		Use:   FlagsetName,
		Short: "Run using CLI flags.",
	}

	for i, mod := range modules {
		meta := &module.Meta{Module: mod}
		metadata[i] = meta

		modArgs := make(module.Args, 0, len(mod.Args())+len(mod.Ops())*argsPerOp)
		modArgs = append(modArgs, mod.Args()...)

		for _, op := range mod.Ops() {
			modArgs = append(modArgs, disableArg(op))
			modArgs = append(modArgs, rateArg(op))
		}

		if err := registerFlags(cmd, strings.ToLower(mod.Name()), modArgs, &required); err != nil {
			return nil, err
		}
	}

	cmd.RunE = func(cmd *cobra.Command, _ []string) error {
		for _, name := range required {
			if !cmd.Flags().Changed(name) {
				return fmt.Errorf("%w: --%s is required", module.ErrArgRequired, name)
			}
		}

		return run(metadata)
	}

	return cmd, nil
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
