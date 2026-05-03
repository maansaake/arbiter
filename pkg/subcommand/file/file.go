// Package file implements support for the 'file' subcommand.
package file

import (
	"github.com/maansaake/arbiter/pkg/module"
	"github.com/maansaake/arbiter/pkg/subcommand"
)

const FlagsetName = "file"

func Parse(_ int, _ module.Modules) ([]*subcommand.Meta, error) {
	return nil, nil
}
