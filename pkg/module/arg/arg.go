package arg

import "errors"

type (
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

var ErrParse = errors.New("error parsing CLI flags")
