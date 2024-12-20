package module

import (
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"
	"time"
)

type (
	Module interface {
		Name() string
		Desc() string
		Args() Args
		Ops() Ops
		Run() error
		Stop() error
	}
	Modules []Module
	Op      struct {
		Name     string
		Desc     string
		Disabled bool
		Do
		Rate uint
	}
	Ops    []*Op
	Do     func() (Result, error)
	Result struct {
		Duration time.Duration
	}
	TypeConstraint interface {
		~int | ~uint | ~float64 | ~string | ~bool
	}
	Arg[T TypeConstraint] struct {
		// The name of the argument. This is used to name CLI and file options when
		// starting the arbiter.
		Name string
		// A description tied to the argument, this will show as a hint in the CLI,
		// and as a comment in the generated file.
		Desc string
		// If set, the arbiter will throw an error on start if the argument was not
		// provided.
		Required bool
		// A pointer to the value of the argument. This is used both to provide a
		// default and it will be populated with a given argument's value. You have
		// to set a value unless a Handler is given.
		Value *T
		// The handler function is called with the parsed argument value if given.
		// You can use a handler and set a default value in combination. A handler
		// can be useful to perform additional parsing or conversion.
		Handler func(v T)
		// A validator function for the argument value.
		Valid Validator[T]
	}
	Args                        []any
	Validator[T TypeConstraint] func(v T) bool
)

var (
	reservedPrefixes = []string{"arbiter", "monitor", "reporter"}

	moduleNameRe *regexp.Regexp
	opNameRe     *regexp.Regexp

	ErrReservedPrefix = errors.New("module name is reserved")
	ErrInvalidName    = errors.New("name is invalid")
	ErrArgParse       = errors.New("failed to parse argument")
	ErrArgRequired    = errors.New("argument is required")
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
