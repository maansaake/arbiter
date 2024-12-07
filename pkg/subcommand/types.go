package subcommand

import (
	"tres-bon.se/arbiter/pkg/module"
	"tres-bon.se/arbiter/pkg/monitor"
)

// A collection type to help coordination between packages without relying on
// weird import patterns. All subcommands make use of this struct to be able
// to do their work, regardless of whether running using a test file or CLI.
type ModuleMeta struct {
	// A module.
	module.Module
	// Monitoring information for the module
	*monitor.ModuleInfo
}
