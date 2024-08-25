package manager

import (
	"tres-bon.se/assure/pkg/monitor"
	"tres-bon.se/assure/pkg/reporter"
	"tres-bon.se/assure/pkg/testmodule"
	"tres-bon.se/assure/pkg/traffic"
)

type manager struct {
	modules []testmodule.Module

	generator *traffic.Generator
	monitor   *monitor.Monitor
	reporter  reporter.Reporter
}

func new(modules []testmodule.Module) *manager {
	return &manager{
		modules:   modules,
		generator: &traffic.Generator{},
		monitor:   &monitor.Monitor{},
		reporter:  &reporter.YAMLReporter{},
	}
}

func (m *manager) Run() error {
	return nil
}

func (m *manager) Stop() error {
	return nil
}
