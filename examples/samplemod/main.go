package main

import (
	"os"

	"tres-bon.se/arbiter"
	samplemod "tres-bon.se/arbiter/examples/samplemod/module"
	"tres-bon.se/arbiter/pkg/module"
)

func main() {
	err := arbiter.Run(module.Modules{samplemod.New()})
	if err != nil {
		os.Exit(1)
	}
}
