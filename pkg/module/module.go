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
		// Name of the module.
		Name() string
		// Desc is a description tied to the module.
		Desc() string
		// Args returns a list of arguments that the module accepts.
		Args() Args
		// Ops returns a list of operations that the module performs.
		Ops() Ops
		// Run is called when the arbiter starts. This can be used to perform any setup required by the module.
		Run() error
		// Stop is called when the arbiter stops. This can be used to perform any cleanup required by the module.
		Stop() error
	}
	// Modules is a list of Module.
	Modules []Module
	// Op represents an operation that a module performs. It contains the name, description, and the function to execute the operation.
	Op struct {
		// Name of the operation. This is used to name CLI and file options when starting the arbiter.
		Name string
		// Desc is a description tied to the operation, this will show as a hint in the CLI, and as a comment in the generated file.
		Desc string
		// Disabled, if set, the operation will not be executed. This can be used to temporarily disable an operation without removing it from the module.
		Disabled bool
		// Do is the function that will be executed for the operation.
		Do
		// Rate is the number of times the operation should be executed per second. If zero, the operation will be executed as fast as possible.
		Rate uint
	}
	// Ops is a list of Op.
	Ops []*Op
	// Do is the function that will be executed for the operation.
	Do func() (Result, error)
	// Result is the result of an operation.
	Result struct {
		Duration time.Duration
	}
	// TypeConstraint is a constraint that allows only certain types for the argument value.
	TypeConstraint interface {
		~int | ~uint | ~float64 | ~string | ~bool
	}
	Arg[T TypeConstraint] struct {
		// Name of the argument. This is used to name CLI and file options when
		// starting the arbiter.
		Name string
		// Desc is a description tied to the argument, this will show as a hint in the CLI,
		// and as a comment in the generated file.
		Desc string
		// Required, if set, the arbiter will throw an error on start if the argument was not
		// provided.
		Required bool
		// Value is a pointer to the value of the argument. This is used both to provide a
		// default and it will be populated with a given argument's value. You have
		// to set a value unless a Handler is given.
		Value *T
		// Handler function is called with the parsed argument value if given.
		// You can use a handler and set a default value in combination. A handler
		// can be useful to perform additional parsing or conversion.
		Handler func(v T)
		// Valid is a validator function for the argument value.
		Valid Validator[T]
	}
	// Args is a list of arguments that a module accepts.
	Args []any
	// Validator is a function that validates the argument value. It should return true if the value is valid, and false otherwise.
	Validator[T TypeConstraint] func(v T) bool
)

var (
	reservedPrefixes = []string{"arbiter", "reporter"} //nolint:gochecknoglobals // constant-like list of reserved names

	moduleNameRe = regexp.MustCompile(moduleNamePattern)
	opNameRe     = regexp.MustCompile(opNamePattern)

	ErrReservedPrefix = errors.New("module name is reserved")
	ErrInvalidName    = errors.New("name is invalid")
	ErrArgParse       = errors.New("failed to parse argument")
	ErrArgRequired    = errors.New("argument is required")
)

const (
	moduleNamePattern = "^[a-z0-9-]+$"
	opNamePattern     = moduleNamePattern
)

// Validate verifies input modules follow the rules, which are:
// - The module is not named using any of the reserved prefixes.
func Validate(modules Modules) error {
	for _, mod := range modules {
		if slices.Contains(reservedPrefixes, strings.ToLower(mod.Name())) {
			return fmt.Errorf("%w: '%s' cannot be used", ErrReservedPrefix, mod.Name())
		}

		if !moduleNameRe.MatchString(mod.Name()) {
			return fmt.Errorf(
				"%w: module name '%s' does not follow pattern '%s'",
				ErrInvalidName,
				mod.Name(),
				moduleNamePattern,
			)
		}

		for _, op := range mod.Ops() {
			if !opNameRe.MatchString(op.Name) {
				return fmt.Errorf(
					"%w: operation name '%s' does not follow pattern '%s'",
					ErrInvalidName,
					op.Name,
					opNameRe,
				)
			}
		}
	}
	return nil
}
