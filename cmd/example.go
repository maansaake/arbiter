package main

import (
	"tres-bon.se/arbiter"
	"tres-bon.se/arbiter/pkg/module"
	samplemodule "tres-bon.se/arbiter/sample"
)

func main() {
	arbiter.Run(module.Modules{samplemodule.NewSampleModule()})
}
