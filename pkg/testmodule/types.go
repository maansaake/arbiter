package testmodule

import (
	"errors"

	args "tres-bon.se/assure/pkg/arguments"
	op "tres-bon.se/assure/pkg/testmodule/operations"
)

var (
	ErrRun = errors.New("failed to run module")
)

type Module interface {
	Config() args.Args
	Operations() op.Operations
	Run() error
	Stop() error
}

type Modules []Module
