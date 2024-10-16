package main

import (
	"os"

	"tres-bon.se/arbiter"
	"tres-bon.se/arbiter/pkg/module"
	samplemodule "tres-bon.se/arbiter/sample"
)

func main() {
	err := arbiter.Run(module.Modules{samplemodule.NewSampleModule()})
	if err != nil {
		os.Exit(1)
	}
}
