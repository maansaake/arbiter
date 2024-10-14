package arbiter

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"tres-bon.se/arbiter/pkg/arg"
	"tres-bon.se/arbiter/pkg/module"
)

const (
	FLAGSET_GEN  = "generate"
	FLAGSET_FILE = "file"
)

func Run(modules module.Modules) error {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "%s [subcommand]\n", os.Args[0])
		fmt.Fprint(flag.CommandLine.Output(), "  subcommands:\n")
		fmt.Fprint(flag.CommandLine.Output(), "    cli      Run using CLI flags.\n")
		fmt.Fprint(flag.CommandLine.Output(), "    generate Generate a test model.\n")
		fmt.Fprint(flag.CommandLine.Output(), "    file     Run from a test model file.\n")
		flag.PrintDefaults()
	}

	// Top level parse
	flag.Parse()

	if len(os.Args) < 2 {
		fmt.Fprint(flag.CommandLine.Output(), "no subcommand given\n")
		flag.Usage()
		os.Exit(1)
	}

	// Check invoked subcommand
	switch os.Args[1] {
	case arg.FLAGSET:
		return handleCli(modules)
	case FLAGSET_GEN:
		return handleGen(modules)
	case FLAGSET_FILE:
		return handleFile(modules)
	default:
		fmt.Fprintf(flag.CommandLine.Output(), "subcommand not found: %s\n", os.Args[1])
		flag.Usage()
		os.Exit(1)
	}

	return nil
}

func handleCli(modules module.Modules) error {
	for _, m := range modules {
		for _, a := range m.Args() {
			arg.Register(m.Name(), a)
		}

		// Add operation args
		for _, op := range m.Ops() {
			arg.RegisterOp(m.Name(), op)
		}
	}

	arg.Parse(os.Args[2:])

	return runModules(modules)
}

func handleGen(modules module.Modules) error {
	var output string
	fs := flag.NewFlagSet(FLAGSET_GEN, flag.ExitOnError)
	fs.StringVar(&output, "output", "./arbiter.yaml", "Output path for the generated test model file.")
	fs.Parse(os.Args[2:])

	// TODO: generate using input modules
	panic("not implemented")

	return nil
}

func handleFile(modules module.Modules) error {
	var path string
	fs := flag.NewFlagSet(FLAGSET_FILE, flag.ExitOnError)
	fs.StringVar(&path, "path", "./arbiter.yaml", "Path to a test model file.")
	fs.Parse(os.Args[2:])

	// TODO: parse and run from file
	panic("not implemented")

	return runModules(modules)
}

func runModules(modules module.Modules) error {
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
