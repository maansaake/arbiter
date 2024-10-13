package arbiter

import (
	"tres-bon.se/arbiter/pkg/arg"
	"tres-bon.se/arbiter/pkg/module"
	"tres-bon.se/arbiter/pkg/monitor"
	"tres-bon.se/arbiter/pkg/report"
)

func Run(modules module.Modules, reporter report.Reporter, monitor monitor.Monitor) error {
	// Extract module args
	args := make(arg.Args, 0)
	for _, m := range modules {
		args = append(args, m.Args()...)
		// Add operation args
		// for _, op := range m.Ops() {
		// args = append(args, makeArg(op))
		// }
	}

	// Args for the monitor

	// Args for the reporter

	if err := arg.ParseArgs(args); err != nil {

	}

	// Start each module

	// Start signal interceptor for SIGINT and SIGTERM

	// Await done channel

	return nil
}
