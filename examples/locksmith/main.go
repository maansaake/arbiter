package main

import (
	"os"

	"tres-bon.se/arbiter"
	"tres-bon.se/arbiter/pkg/module"
)

func main() {
	err := arbiter.Run(module.Modules{newLocksmithModule()})
	if err != nil {
		os.Exit(1)
	}
}
