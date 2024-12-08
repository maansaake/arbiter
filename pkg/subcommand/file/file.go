// The file package implements support for the 'file' subcommand.
package file

import (
	"tres-bon.se/arbiter/pkg/module"
	"tres-bon.se/arbiter/pkg/subcommand"
)

const FLAGSET = "file"

func Parse(subcommandIndex int, _ module.Modules) ([]*subcommand.Meta, error) {
	return nil, nil
}
