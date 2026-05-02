package subcommand

import (
	"tres-bon.se/arbiter/pkg/module"
)

// Meta is a collection type to help coordination between packages without relying on
// weird import patterns. All subcommands make use of this struct to be able
// to do their work, regardless of whether running using a test file or CLI.
type Meta struct {
	// A module.
	module.Module
}

type Metadata []*Meta
