package main

import (
	"os"

	"github.com/maansaake/arbiter"
	samplemod "github.com/maansaake/arbiter/examples/samplemod/module"
	"github.com/maansaake/arbiter/pkg/module"
)

func main() {
	err := arbiter.Run(module.Modules{samplemod.New()})
	if err != nil {
		os.Exit(1)
	}
}
