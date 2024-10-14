package arbiter

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"tres-bon.se/arbiter/pkg/arg"
	"tres-bon.se/arbiter/pkg/module"
)

func Run(modules module.Modules) error {
	// Register args if first arg is not a path to a traffic model file, check arg does not start with '-'
	log.Println(os.Args[1])

	// TODO: generate traffic model if first arg is ...

	for _, m := range modules {
		for _, a := range m.Args() {
			arg.Register(a)
		}

		// Add operation args
		for _, op := range m.Ops() {
			arg.RegisterOp(m.Name(), op)
		}
	}

	// TODO: args for the monitor

	// TODO: args for the reporter

	// Parse args
	arg.Parse()

	// Start each module
	for _, m := range modules {
		log.Printf("starting module '%s'\n", m.Name())
		if err := m.Run(); err != nil {
			log.Fatalf("failed to start module '%s': %s", m.Name(), err.Error())
		}
	}

	// Start signal interceptor for SIGINT and SIGTERM
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	// TODO: start traffic

	// Await done channel
	log.Println("awaiting stop signal")
	<-signals

	log.Println("got stop signal")
	for _, m := range modules {
		if err := m.Stop(); err != nil {
			log.Printf("stop error when stopping module '%s': %s\n", m.Name(), err.Error())
		}
	}

	return nil
}
