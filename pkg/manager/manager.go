package manager

import (
	"errors"

	"github.com/go-logr/logr"
	"tres-bon.se/assure/pkg/log"
	"tres-bon.se/assure/pkg/testmodule"
)

// Run starts the Assure manager with the given modules
func Run(modules testmodule.Modules) error {
	mgr := new(modules)

	log.Set(logr.New(log.NewSink(&log.Options{
		Verbosity:          50,
		VerbosityFieldName: "v",
		NameFieldName:      "",
		Console:            true,
		Caller:             true,
	})))

	log.Info("hello there")
	log.Error(errors.New("an error"), "here's an error")
	log.V(2).Info("here's a more verbose message")

	if err := mgr.Run(); err != nil {
		return err
	}

	return nil
}
