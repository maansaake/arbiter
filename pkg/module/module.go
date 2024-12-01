package module

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"tres-bon.se/arbiter/pkg/arg"
	"tres-bon.se/arbiter/pkg/module/op"
)

var (
	reservedPrefixes = []string{"arbiter", "monitor", "reporter"}

	ErrReservedPrefix = errors.New("")
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

type Meta struct {
	Module
	*Monitoring
}

type Monitoring struct {
	PID            int
	LogFile        string
	MetricEndpoint string
}

func NewMeta(module Module) *Meta {
	return &Meta{Module: module, Monitoring: &Monitoring{}}
}

// Verifies input modules follow the rules, which are:
// - The module is not named using any of the reserved prefixes
func Validate(modules Modules) error {
	for _, mod := range modules {
		if slices.Contains(reservedPrefixes, strings.ToLower(mod.Name())) {
			return fmt.Errorf("%w: modules name '%s' cannot be used, it's reserved", ErrReservedPrefix, mod.Name())
		}
	}
	return nil
}
