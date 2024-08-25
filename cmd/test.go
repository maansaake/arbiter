package main

import (
	"tres-bon.se/assure/pkg/manager"
	"tres-bon.se/assure/pkg/testmodule"
)

func main() {
	mgr := manager.New(testmodule.Modules{})

	if err := mgr.Run(); err != nil {
	}
}
