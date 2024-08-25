package main

import (
	"tres-bon.se/assure/pkg/manager"
	"tres-bon.se/assure/pkg/testmodule"
)

func main() {
	if err := manager.Run(testmodule.Modules{}); err != nil {
	}
}
