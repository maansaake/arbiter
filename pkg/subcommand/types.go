package subcommand

import (
	"tres-bon.se/arbiter/pkg/module"
	"tres-bon.se/arbiter/pkg/monitor"
)

type ModuleMeta struct {
	module.Module
	*monitor.ModuleInfo
}
