package main

import (
	"os"

	"tres-bon.se/arbiter"
	locksmith "tres-bon.se/arbiter/examples/locksmith/module"
	"tres-bon.se/arbiter/pkg/module"
)

func main() {
	err := arbiter.Run(module.Modules{locksmith.NewLocksmithModule()})
	if err != nil {
		os.Exit(1)
	}
}
