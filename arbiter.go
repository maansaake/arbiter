package arbiter

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"slices"
	"syscall"
	"time"

	"tres-bon.se/arbiter/pkg/arg"
	"tres-bon.se/arbiter/pkg/module"
)

const (
	FLAGSET_CLI  = arg.FLAGSET
	FLAGSET_GEN  = "generate"
	FLAGSET_FILE = "file"
)

var (
	duration time.Duration

	subcommands     = []string{FLAGSET_CLI, FLAGSET_GEN, FLAGSET_FILE}
	subcommandIndex = -1
)

// Runs the Arbiter. Blocks until SIGINT, SIGTERM or when the test duration
// runs out (1 hour default).
func Run(modules module.Modules) error {
	flag.CommandLine.SetOutput(os.Stdout)
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "%s [subcommand]\n\n", os.Args[0])
		fmt.Fprint(flag.CommandLine.Output(), "subcommands:\n")
		fmt.Fprint(flag.CommandLine.Output(), "  cli      Run using CLI flags.\n")
		fmt.Fprint(flag.CommandLine.Output(), "  generate Generate a test model file.\n")
		fmt.Fprint(flag.CommandLine.Output(), "  file     Run from a test model file.\n")
		fmt.Fprint(flag.CommandLine.Output(), "\n")
		fmt.Fprint(flag.CommandLine.Output(), "global flags:\n")
		flag.PrintDefaults()
	}

	// Global flags
	flag.DurationVar(&duration, "duration", 3600*time.Second, "The duration of the test run.")

	// To trigger on --help and parse global flags
	flag.Parse()

	subcommandIndex = slices.IndexFunc(os.Args, func(e string) bool {
		return slices.Contains(subcommands, e)
	})
	if subcommandIndex == -1 {
		flag.CommandLine.SetOutput(os.Stderr)
		fmt.Fprint(flag.CommandLine.Output(), "no subcommand given\n")
		flag.Usage()
		os.Exit(1)
	}

	// Check invoked subcommand
	switch os.Args[subcommandIndex] {
	case arg.FLAGSET:
		return handleCli(modules)
	case FLAGSET_GEN:
		return handleGen(modules)
	case FLAGSET_FILE:
		return handleFile(modules)
	default:
		flag.CommandLine.SetOutput(os.Stderr)
		fmt.Fprintf(flag.CommandLine.Output(), "subcommand not found: %s\n", os.Args[1])
		flag.Usage()
		os.Exit(1)
	}

	return nil
}

// Handle the CLI subcommand call, registering cmd line flags for both module
// arguments and their operations and parsing them. Runs the modules with the
// input flags.
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

	arg.Parse(os.Args[subcommandIndex+1:])

	return run(modules)
}

// Handles the generate subcommand. Generates a test model file based on the
// input modules.
func handleGen(_ module.Modules) error {
	var output string
	fs := flag.NewFlagSet(FLAGSET_GEN, flag.ExitOnError)
	fs.StringVar(&output, "output", "./arbiter.yaml", "Output path for the generated test model file.")
	err := fs.Parse(os.Args[subcommandIndex+1:])
	if err != nil {
		fs.SetOutput(os.Stderr)
		fs.Usage()
		os.Exit(1)
	}

	// TODO: generate using input modules
	panic("not implemented")
}

// Handles the file subcommand, parsing the input test model file and running
// the modules with the file's settings.
func handleFile(_ module.Modules) error {
	var path string
	fs := flag.NewFlagSet(FLAGSET_FILE, flag.ExitOnError)
	fs.StringVar(&path, "path", "./arbiter.yaml", "Path to a test model file.")
	err := fs.Parse(os.Args[subcommandIndex+1:])
	if err != nil {
		fs.SetOutput(os.Stderr)
		fs.Usage()
		os.Exit(1)
	}

	// TODO: parse and run from file
	panic("not implemented")
}

// Runs the input modules and starts generating traffic. Creates a traffic
// model based on the modules opertation settings. Aborts on SIGINT, SIGTERM
// or when the test duration runs out. Will immediately exit if any module
// returns an error from its call to Run().
func run(modules module.Modules) error {
	// Start each module, exit on error
	for _, m := range modules {
		log.Printf("starting module '%s'\n", m.Name())
		if err := m.Run(); err != nil {
			log.Fatalf("failed to start module '%s': %s", m.Name(), err.Error())
		}
	}

	// TODO: start traffic

	// TODO: await done channel
	// Start signal interceptor for SIGINT and SIGTERM
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
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
