package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/maansaake/arbiter"
	samplemod "github.com/maansaake/arbiter/examples/samplemod/module"
	"github.com/maansaake/arbiter/pkg/module"
)

func main() {
	err := arbiter.Run(module.Modules{samplemod.New()})
	if err == nil {
		os.Exit(0)
	}

	// This inspection allows us to provide a more specific error message when the error is related to stopping traffic or modules,
	if errors.Is(err, arbiter.ErrStopping) {
		fmt.Fprintf(os.Stderr, "Arbiter stopped with error: %v\n\n", err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "Error running the Arbiter: %v\n\n", err)
	os.Exit(1)
}
