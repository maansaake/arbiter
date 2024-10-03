package module

import (
	"tres-bon.se/arbiter/pkg/arg"
	"tres-bon.se/arbiter/pkg/module/op"
)

type Module interface {
	Name() string
	Desc() string
	Args() arg.Args
	Ops() op.Ops
	Run() error
	Stop() error
}

type Modules []Module
