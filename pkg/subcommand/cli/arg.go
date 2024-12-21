package cli

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"slices"
	"strconv"

	"tres-bon.se/arbiter/pkg/module"
)

const FlagsetName = "cli"

var (
	flagset  = flag.NewFlagSet(FlagsetName, flag.ExitOnError)
	required []string

	ErrNilPtr       = errors.New("Arg.Value must not be a nil pointer")
	ErrRequiredBool = errors.New("a boolean arg cannot be marked required")
	ErrInvalid      = errors.New("validator failed")
	ErrType         = errors.New("unsupported type")
)

func Register(prefix string, args module.Args) error {
	errs := make([]error, 0)
	for _, arg := range args {
		errs = append(errs, register(prefix, arg))
	}

	return errors.Join(errs...)
}

// Parses command line input.
func ParseArgs(args []string) error {
	// Ignore error since we're using ExitOnError.
	_ = flagset.Parse(args)

	if len(required) > 0 {
		for _, an := range required {
			fmt.Fprintf(os.Stderr, "missing required argument -%s\n", an)
		}
		flagset.SetOutput(os.Stderr)
		flagset.Usage()
		return fmt.Errorf("%w: %d required flags have been missed", module.ErrArgRequired, len(required))
	}

	return nil
}

// Register the argument with the given prefix, resulting in a command line
// flag: prefix.<argument.Name>
func register(prefix string, argument any) error {
	switch typedArgument := argument.(type) {
	case *module.Arg[int]:
		return registerInt(prefix, typedArgument)
	case *module.Arg[uint]:
		return registerUint(prefix, typedArgument)
	case *module.Arg[float64]:
		return registerFloat(prefix, typedArgument)
	case *module.Arg[string]:
		return registerString(prefix, typedArgument)
	case *module.Arg[bool]:
		return registerBool(prefix, typedArgument)
	}
	// This is basically a type constraint mismatch.
	return ErrType
}

func registerInt(prefix string, arg *module.Arg[int]) error {
	if err := verifyArgValue(arg); err != nil {
		return err
	}

	if arg.Required {
		required = append(required, argPath(prefix, arg))
	}

	flagset.Func(argPath(prefix, arg), arg.Desc, intHandler(prefix, arg))
	return nil
}

func intHandler(prefix string, arg *module.Arg[int]) func(string) error {
	return func(val string) error {
		iv, err := strconv.ParseInt(val, 10, 0)
		if err != nil {
			return err
		}
		*arg.Value = int(iv)

		return generalHandler(prefix, arg)
	}
}

func registerUint(prefix string, arg *module.Arg[uint]) error {
	if err := verifyArgValue(arg); err != nil {
		return err
	}

	if arg.Required {
		required = append(required, argPath(prefix, arg))
	}

	flagset.Func(argPath(prefix, arg), arg.Desc, uintHandler(prefix, arg))
	return nil
}

func uintHandler(prefix string, arg *module.Arg[uint]) func(string) error {
	return func(val string) error {
		iv, err := strconv.ParseUint(val, 10, 0)
		if err != nil {
			return err
		}
		*arg.Value = uint(iv)

		return generalHandler(prefix, arg)
	}
}

func registerFloat(prefix string, arg *module.Arg[float64]) error {
	if err := verifyArgValue(arg); err != nil {
		return err
	}

	if arg.Required {
		required = append(required, argPath(prefix, arg))
	}

	flagset.Func(argPath(prefix, arg), arg.Desc, floatHandler(prefix, arg))
	return nil
}

func floatHandler(prefix string, arg *module.Arg[float64]) func(string) error {
	return func(val string) error {
		fv, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return err
		}
		*arg.Value = fv

		return generalHandler(prefix, arg)
	}
}

func registerString(prefix string, arg *module.Arg[string]) error {
	if err := verifyArgValue(arg); err != nil {
		return err
	}

	if arg.Required {
		required = append(required, argPath(prefix, arg))
	}

	flagset.Func(argPath(prefix, arg), arg.Desc, stringHandler(prefix, arg))

	return nil
}

func stringHandler(prefix string, arg *module.Arg[string]) func(string) error {
	return func(val string) error {
		*arg.Value = val
		return generalHandler(prefix, arg)
	}
}

func registerBool(prefix string, arg *module.Arg[bool]) error {
	if err := verifyArgValue(arg); err != nil {
		return err
	}

	if arg.Required {
		return fmt.Errorf("%w: '%s'", ErrRequiredBool, argPath(prefix, arg))
	}

	flagset.BoolVar(arg.Value, argPath(prefix, arg), *arg.Value, arg.Desc)
	return nil
}

func verifyArgValue[T module.TypeConstraint](arg *module.Arg[T]) error {
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
func generalHandler[T module.TypeConstraint](prefix string, arg *module.Arg[T]) error {
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

func argPath[T module.TypeConstraint](prefix string, arg *module.Arg[T]) string {
	return fmt.Sprintf("%s.%s", prefix, arg.Name)
}
