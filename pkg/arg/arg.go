package arg

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"slices"
	"strconv"
)

type (
	Arg[T any] struct {
		Name     string
		Desc     string
		Required bool
		Value    *T
		Valid    func(v T) bool
	}
	Args []any
)

const FLAGSET = "cli"

var (
	flagset  = flag.NewFlagSet(FLAGSET, flag.ContinueOnError)
	required []string

	ErrNilPtr       = errors.New("Arg.Value must not be a nil pointer")
	ErrRequiredBool = errors.New("a boolean arg cannot be marked required")
	ErrInvalid      = errors.New("invalid value")
	ErrParseError   = errors.New("error parsing CLI flags")
)

func Register(prefix string, args Args) error {
	errs := make([]error, 0)
	for _, arg := range args {
		errs = append(errs, register(prefix, arg))
	}

	return errors.Join(errs...)
}

// Parses command line input.
func Parse(args []string) error {
	err := flagset.Parse(args)
	if err != nil {
		// Usage() is called directly from the flagset.Parse call in this case.
		return err
	}

	if len(required) > 0 {
		for _, an := range required {
			fmt.Fprintf(os.Stderr, "missing required argument -%s\n", an)
		}
		flagset.SetOutput(os.Stderr)
		flagset.Usage()
		return fmt.Errorf("%w: %d required flags have been missed", ErrParseError, len(required))
	}

	return nil
}

// Register the argument with the given prefix, resulting in a command line
// flag: prefix.<argument.Name>
func register(prefix string, argument any) error {
	switch typedArgument := argument.(type) {
	case *Arg[int]:
		return registerInt(prefix, typedArgument)
	case *Arg[uint]:
		return registerUint(prefix, typedArgument)
	case *Arg[float64]:
		return registerFloat(prefix, typedArgument)
	case *Arg[string]:
		return registerString(prefix, typedArgument)
	case *Arg[bool]:
		return registerBool(prefix, typedArgument)
	default:
		panic("argument type not implemented")
	}
}

func registerInt(prefix string, arg *Arg[int]) error {
	if err := verifyArgValue(arg); err != nil {
		return err
	}

	if arg.Required {
		required = append(required, argPath(prefix, arg))
	}

	flagset.Func(argPath(prefix, arg), arg.Desc, intHandler(prefix, arg))
	return nil
}

func intHandler(prefix string, arg *Arg[int]) func(string) error {
	return func(val string) error {
		iv, err := strconv.ParseInt(val, 10, 0)
		if err != nil {
			return err
		}
		*arg.Value = int(iv)

		return generalHandler(prefix, arg)
	}
}

func registerUint(prefix string, arg *Arg[uint]) error {
	if err := verifyArgValue(arg); err != nil {
		return err
	}

	if arg.Required {
		required = append(required, argPath(prefix, arg))
	}

	flagset.Func(argPath(prefix, arg), arg.Desc, uintHandler(prefix, arg))
	return nil
}

func uintHandler(prefix string, arg *Arg[uint]) func(string) error {
	return func(val string) error {
		iv, err := strconv.ParseUint(val, 10, 0)
		if err != nil {
			return err
		}
		*arg.Value = uint(iv)

		return generalHandler(prefix, arg)
	}
}

func registerFloat(prefix string, arg *Arg[float64]) error {
	if err := verifyArgValue(arg); err != nil {
		return err
	}

	if arg.Required {
		required = append(required, argPath(prefix, arg))
	}

	flagset.Func(argPath(prefix, arg), arg.Desc, floatHandler(prefix, arg))
	return nil
}

func floatHandler(prefix string, arg *Arg[float64]) func(string) error {
	return func(val string) error {
		fv, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return err
		}
		*arg.Value = fv

		return generalHandler(prefix, arg)
	}
}

func registerString(prefix string, arg *Arg[string]) error {
	if err := verifyArgValue(arg); err != nil {
		return err
	}

	if arg.Required {
		required = append(required, argPath(prefix, arg))
	}

	flagset.Func(argPath(prefix, arg), arg.Desc, stringHandler(prefix, arg))

	return nil
}

func stringHandler(prefix string, arg *Arg[string]) func(string) error {
	return func(val string) error {
		*arg.Value = val
		return generalHandler(prefix, arg)
	}
}

func registerBool(prefix string, arg *Arg[bool]) error {
	if err := verifyArgValue(arg); err != nil {
		return err
	}

	if arg.Required {
		return fmt.Errorf("%w: '%s'", ErrRequiredBool, argPath(prefix, arg))
	}

	flagset.BoolVar(arg.Value, argPath(prefix, arg), *arg.Value, arg.Desc)
	return nil
}

func verifyArgValue[T any](arg *Arg[T]) error {
	if arg.Value == nil {
		return fmt.Errorf("%w: '%s'", ErrNilPtr, arg.Name)
	}
	return nil
}

// Handle required, validation and all other actions.
func generalHandler[T any](prefix string, arg *Arg[T]) error {
	if arg.Required {
		// Find and pop arg from required slice
		for i, an := range required {
			if an == argPath(prefix, arg) {
				required = slices.Delete(required, i, i+1)
			}
		}
	}

	if arg.Valid != nil && !arg.Valid(*arg.Value) {
		return fmt.Errorf("%w: argument '%s' has invalid value '%v'", ErrInvalid, argPath(prefix, arg), arg.Value)
	}

	return nil
}

func argPath[T any](prefix string, arg *Arg[T]) string {
	return fmt.Sprintf("%s.%s", prefix, arg.Name)
}
