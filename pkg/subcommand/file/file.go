// Package file implements support for the 'file' subcommand.
package file

import (
	"tres-bon.se/arbiter/pkg/module"
	"tres-bon.se/arbiter/pkg/subcommand"
)

const FlagsetName = "file"

func Parse(_ int, _ module.Modules) ([]*subcommand.Meta, error) {
	return nil, nil
}
