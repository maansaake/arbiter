package module

import (
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"

	"tres-bon.se/arbiter/pkg/module/arg"
	"tres-bon.se/arbiter/pkg/module/op"
)

type (
	Module interface {
		Name() string
		Desc() string
		Args() arg.Args
		Ops() op.Ops
		Run() error
		Stop() error
	}
	Modules []Module
)

var (
	reservedPrefixes = []string{"arbiter", "monitor", "reporter"}

	moduleNameRe *regexp.Regexp
	opNameRe     *regexp.Regexp

	ErrReservedPrefix = errors.New("module name is reserved")
	ErrInvalidName    = errors.New("name is invalid")
)

const (
	moduleNamePattern = "^[a-z0-9-]+$"
	opNamePattern     = moduleNamePattern
)

func init() {
	moduleNameRe = regexp.MustCompile(moduleNamePattern)
	opNameRe = regexp.MustCompile(opNamePattern)
}

// Verifies input modules follow the rules, which are:
// - The module is not named using any of the reserved prefixes
func Validate(modules Modules) error {
	for _, mod := range modules {
		if slices.Contains(reservedPrefixes, strings.ToLower(mod.Name())) {
			return fmt.Errorf("%w: '%s' cannot be used", ErrReservedPrefix, mod.Name())
		}

		if !moduleNameRe.MatchString(mod.Name()) {
			return fmt.Errorf("%w: module name '%s' does not follow pattern '%s'", ErrInvalidName, mod.Name(), moduleNamePattern)
		}

		for _, op := range mod.Ops() {
			if !opNameRe.MatchString(op.Name) {
				return fmt.Errorf("%w: operation name '%s' does not follow pattern '%s'", ErrInvalidName, op.Name, opNameRe)
			}
		}
	}
	return nil
}
