package arg

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"slices"
	"strconv"

	"tres-bon.se/arbiter/pkg/module/op"
)

type Arg[T any] struct {
	Name     string
	Desc     string
	Required bool
	Value    T
	Valid    Validator[T]
}

type Validator[T any] func(v T) bool

type Args []any

var (
	ErrInvalid = errors.New("Invalid value")

	required []string
)

func Register(arg any) {
	switch ta := arg.(type) {
	case *Arg[int]:
		registerInt(ta)
	case *Arg[float64]:
		registerFloat(ta)
	case *Arg[string]:
		registerString(ta)
	case *Arg[bool]:
		registerBool(ta)
	default:
		panic("not implemented")
	}
}

func RegisterOp(op *op.Op) {

}

func Parse() {
	flag.Parse()

	// Print all problems before exiting.
	problem := false

	if len(required) > 0 {
		problem = true
		for _, an := range required {
			fmt.Fprintf(os.Stderr, "missing required argument -%s\n", an)
		}
	}

	if problem {
		flag.Usage()
		os.Exit(1)
	}
}

func registerInt(arg *Arg[int]) {
	if arg.Required {
		required = append(required, arg.Name)
	}

	flag.Func(arg.Name, arg.Desc, handleInt(arg))
}

func handleInt(arg *Arg[int]) func(string) error {
	return func(val string) error {
		iv, err := strconv.ParseInt(val, 10, 0)
		if err != nil {
			return err
		}

		arg.Value = int(iv)

		return validationHandler(arg)
	}
}

func registerFloat(arg *Arg[float64]) {
	if arg.Required {
		required = append(required, arg.Name)
	}

	flag.Func(arg.Name, arg.Desc, handleFloat(arg))
}

func handleFloat(arg *Arg[float64]) func(string) error {
	return func(val string) error {
		fv, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return err
		}

		arg.Value = fv

		return validationHandler(arg)
	}
}

func registerString(arg *Arg[string]) {
	if arg.Required {
		required = append(required, arg.Name)
	}

	flag.Func(arg.Name, arg.Desc, handleString(arg))
}

func handleString(arg *Arg[string]) func(string) error {
	return func(val string) error {
		arg.Value = val
		return validationHandler(arg)
	}
}

func registerBool(arg *Arg[bool]) {
	if arg.Required {
		required = append(required, arg.Name)
	}

	flag.Func(arg.Name, arg.Desc, handleBool(arg))
}

func handleBool(arg *Arg[bool]) func(string) error {
	return func(val string) error {
		b, err := strconv.ParseBool(val)
		if err != nil {
			return err
		}

		arg.Value = b

		return validationHandler(arg)
	}
}

func validationHandler[T any](arg *Arg[T]) error {
	if arg.Required {
		// Find and pop arg from required slice
		for i, an := range required {
			if an == arg.Name {
				required = slices.Delete(required, i, i+1)
			}
		}
	}

	if !arg.Valid(arg.Value) {
		return fmt.Errorf("%w: argument '%s' has invalid value '%v'", ErrInvalid, arg.Name, arg.Value)
	}

	return nil
}
