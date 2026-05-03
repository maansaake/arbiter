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

	if errors.Is(err, arbiter.ErrParsingFailed) {
		fmt.Fprintf(os.Stderr, "Error running Arbiter: %v\n\n", err)
		arbiter.Usage()
	}

	os.Exit(1)
}
