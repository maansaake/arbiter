package manager

import (
	"tres-bon.se/assure/pkg/monitor"
	"tres-bon.se/assure/pkg/reporter"
	"tres-bon.se/assure/pkg/testmodule"
	"tres-bon.se/assure/pkg/traffic"
)

type Manager interface {
	Run() error
	Monitor(*monitor.Monitor) Manager
	Reporter(reporter.Reporter) Manager
	Stop() error
}

type managerImpl struct {
	modules []testmodule.Module

	generator *traffic.Generator
	monitor   *monitor.Monitor
	reporter  reporter.Reporter
}

func New(modules []testmodule.Module) Manager {
	return &managerImpl{
		modules:   modules,
		generator: &traffic.Generator{},
		monitor:   &monitor.Monitor{},
		reporter:  &reporter.YAMLReporter{},
	}
}

func (m *managerImpl) Monitor(newMonitor *monitor.Monitor) Manager {
	m.monitor = newMonitor
	return m
}

func (m *managerImpl) Reporter(newReporter reporter.Reporter) Manager {
	m.reporter = newReporter
	return m
}

func (m *managerImpl) Run() error {
	return nil
}

func (m *managerImpl) Stop() error {
	return nil
}
