package arbiter

import (
	"tres-bon.se/arbiter/pkg/arg"
	"tres-bon.se/arbiter/pkg/module"
	"tres-bon.se/arbiter/pkg/monitor"
	"tres-bon.se/arbiter/pkg/report"
)

func Run(modules module.Modules, reporter report.Reporter, monitor monitor.Monitor) error {
	for _, m := range modules {
		for _, a := range m.Args() {
			arg.Register(a)
		}

		// Add operation args
		for _, op := range m.Ops() {
			arg.RegisterOp(m.Name(), op)
		}
	}

	// Args for the monitor

	// Args for the reporter

	// Parse args
	arg.Parse()

	// Start each module

	// Start signal interceptor for SIGINT and SIGTERM

	// Await done channel

	return nil
}
