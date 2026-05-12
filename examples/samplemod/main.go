package main

import (
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

	fmt.Fprintf(os.Stderr, "Error running the Arbiter: %v\n\n", err)
	os.Exit(1)
}
