package operations

import (
	args "tres-bon.se/assure/pkg/arguments"
)

type Operation struct {
	Name string
	Args args.Args
	Do   OperationFunc
}

type Operations []*Operation

type OperationFunc func(keysAndValues ...any) error
