package cli

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"slices"
	"strconv"

	"tres-bon.se/arbiter/pkg/module/arg"
)

const FLAGSET = "cli"

var (
	flagset  = flag.NewFlagSet(FLAGSET, flag.ContinueOnError)
	required []string

	ErrNilPtr       = errors.New("Arg.Value must not be a nil pointer")
	ErrRequiredBool = errors.New("a boolean arg cannot be marked required")
	ErrInvalid      = errors.New("validator failed")
	ErrType         = errors.New("unsupported type")
)

func Register(prefix string, args arg.Args) error {
	errs := make([]error, 0)
	for _, arg := range args {
		errs = append(errs, register(prefix, arg))
	}

	return errors.Join(errs...)
}

// Parses command line input.
func ParseArgs(args []string) error {
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
		return fmt.Errorf("%w: %d required flags have been missed", arg.ErrParse, len(required))
	}

	return nil
}

// Register the argument with the given prefix, resulting in a command line
// flag: prefix.<argument.Name>
func register(prefix string, argument any) error {
	switch typedArgument := argument.(type) {
	case *arg.Arg[int]:
		return registerInt(prefix, typedArgument)
	case *arg.Arg[uint]:
		return registerUint(prefix, typedArgument)
	case *arg.Arg[float64]:
		return registerFloat(prefix, typedArgument)
	case *arg.Arg[string]:
		return registerString(prefix, typedArgument)
	case *arg.Arg[bool]:
		return registerBool(prefix, typedArgument)
	}
	// This is basically a type constraint mismatch.
	return ErrType
}

func registerInt(prefix string, arg *arg.Arg[int]) error {
	if err := verifyArgValue(arg); err != nil {
		return err
	}

	if arg.Required {
		required = append(required, argPath(prefix, arg))
	}

	flagset.Func(argPath(prefix, arg), arg.Desc, intHandler(prefix, arg))
	return nil
}

func intHandler(prefix string, arg *arg.Arg[int]) func(string) error {
	return func(val string) error {
		iv, err := strconv.ParseInt(val, 10, 0)
		if err != nil {
			return err
		}
		*arg.Value = int(iv)

		return generalHandler(prefix, arg)
	}
}

func registerUint(prefix string, arg *arg.Arg[uint]) error {
	if err := verifyArgValue(arg); err != nil {
		return err
	}

	if arg.Required {
		required = append(required, argPath(prefix, arg))
	}

	flagset.Func(argPath(prefix, arg), arg.Desc, uintHandler(prefix, arg))
	return nil
}

func uintHandler(prefix string, arg *arg.Arg[uint]) func(string) error {
	return func(val string) error {
		iv, err := strconv.ParseUint(val, 10, 0)
		if err != nil {
			return err
		}
		*arg.Value = uint(iv)

		return generalHandler(prefix, arg)
	}
}

func registerFloat(prefix string, arg *arg.Arg[float64]) error {
	if err := verifyArgValue(arg); err != nil {
		return err
	}

	if arg.Required {
		required = append(required, argPath(prefix, arg))
	}

	flagset.Func(argPath(prefix, arg), arg.Desc, floatHandler(prefix, arg))
	return nil
}

func floatHandler(prefix string, arg *arg.Arg[float64]) func(string) error {
	return func(val string) error {
		fv, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return err
		}
		*arg.Value = fv

		return generalHandler(prefix, arg)
	}
}

func registerString(prefix string, arg *arg.Arg[string]) error {
	if err := verifyArgValue(arg); err != nil {
		return err
	}

	if arg.Required {
		required = append(required, argPath(prefix, arg))
	}

	flagset.Func(argPath(prefix, arg), arg.Desc, stringHandler(prefix, arg))

	return nil
}

func stringHandler(prefix string, arg *arg.Arg[string]) func(string) error {
	return func(val string) error {
		*arg.Value = val
		return generalHandler(prefix, arg)
	}
}

func registerBool(prefix string, arg *arg.Arg[bool]) error {
	if err := verifyArgValue(arg); err != nil {
		return err
	}

	if arg.Required {
		return fmt.Errorf("%w: '%s'", ErrRequiredBool, argPath(prefix, arg))
	}

	flagset.BoolVar(arg.Value, argPath(prefix, arg), *arg.Value, arg.Desc)
	return nil
}

func verifyArgValue[T arg.TypeConstraint](arg *arg.Arg[T]) error {
	if arg.Handler == nil && arg.Value == nil {
		return fmt.Errorf("%w: '%s'", ErrNilPtr, arg.Name)
	} else if arg.Value == nil {
		// For a Handler arg not having to declare a Value, simplifies things a
		// bit.
		arg.Value = new(T)
	}
	return nil
}

// Handle required, validation and all other actions.
func generalHandler[T arg.TypeConstraint](prefix string, arg *arg.Arg[T]) error {
	if arg.Required {
		// Find and pop arg from required slice
		for i, an := range required {
			if an == argPath(prefix, arg) {
				required = slices.Delete(required, i, i+1)
			}
		}
	}

	if arg.Valid != nil && !arg.Valid(*arg.Value) {
		return ErrInvalid
	}

	if arg.Handler != nil {
		arg.Handler(*arg.Value)
	}

	return nil
}

func argPath[T arg.TypeConstraint](prefix string, arg *arg.Arg[T]) string {
	return fmt.Sprintf("%s.%s", prefix, arg.Name)
}
