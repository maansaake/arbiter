package main

import (
	"tres-bon.se/arbiter"
	"tres-bon.se/arbiter/pkg/module"
)

func main() {
	if err := arbiter.Run(module.Modules{}); err != nil {
	}
}
