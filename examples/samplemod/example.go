package main

import (
	"os"

	"tres-bon.se/arbiter"
	samplemodule "tres-bon.se/arbiter/examples/samplemod/sample"
	"tres-bon.se/arbiter/pkg/module"
)

func main() {
	err := arbiter.Run(module.Modules{samplemodule.NewSampleModule()})
	if err != nil {
		os.Exit(1)
	}
}
