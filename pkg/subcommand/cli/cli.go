// The cli package implements support for the 'cli' subcommand.
package cli

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"slices"
	"strconv"

	"tres-bon.se/arbiter/pkg/arg"
	"tres-bon.se/arbiter/pkg/module"
	"tres-bon.se/arbiter/pkg/module/op"
)

const FLAGSET = "cli"

var (
	flagset  = flag.NewFlagSet(FLAGSET, flag.ContinueOnError)
	required []string

	ErrInvalid    = errors.New("invalid value")
	ErrParseError = errors.New("error parsing CLI flags")
)

// Parse command line arguments for the input modules and populate args and
// operations with parsed values.
func Parse(subcommandIndex int, modules module.Modules) error {
	for _, m := range modules {
		for _, a := range m.Args() {
			register(m.Name(), a)
		}

		// Add operation args
		for _, op := range m.Ops() {
			registerOp(m.Name(), op)
		}
	}

	return parse(os.Args[subcommandIndex+1:])
}

func parse(args []string) error {
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
func register(prefix string, argument any) {
	switch typedArgument := argument.(type) {
	case *arg.Arg[int]:
		registerInt(prefix, typedArgument)
	case *arg.Arg[uint]:
		registerUint(prefix, typedArgument)
	case *arg.Arg[float64]:
		registerFloat(prefix, typedArgument)
	case *arg.Arg[string]:
		registerString(prefix, typedArgument)
	case *arg.Arg[bool]:
		registerBool(prefix, typedArgument)
	default:
		panic("argument type not implemented")
	}
}

// Register the operation arguments with the given prefix, resulting in
// command line flags:
// - prefix.<operation.Name>.disable
// - prefix.<operation.Name>.rate
func registerOp(prefix string, op *op.Op) {
	// Disable
	registerBool(prefix, &arg.Arg[bool]{
		Name:  disableName(op),
		Desc:  disableDesc(op),
		Value: &op.Disabled,
	})
	// Rate
	registerUint(prefix, &arg.Arg[uint]{
		Name:  rateName(op),
		Desc:  rateDesc(op),
		Value: &op.Rate,
	})
}

func disableName(op *op.Op) string {
	return fmt.Sprintf("%s.disable", op.Name)
}

func disableDesc(op *op.Op) string {
	return fmt.Sprintf("Disable %s.", op.Name)
}

func rateName(op *op.Op) string {
	return fmt.Sprintf("%s.rate", op.Name)
}

func rateDesc(op *op.Op) string {
	return fmt.Sprintf("Rate of %s per minute.", op.Name)
}

func registerInt(prefix string, arg *arg.Arg[int]) {
	verifyArgValue(arg)

	if arg.Required {
		required = append(required, argPath(prefix, arg))
	}

	flagset.Func(argPath(prefix, arg), arg.Desc, intHandler(prefix, arg))
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

func registerUint(prefix string, arg *arg.Arg[uint]) {
	verifyArgValue(arg)

	if arg.Required {
		required = append(required, argPath(prefix, arg))
	}

	flagset.Func(argPath(prefix, arg), arg.Desc, uintHandler(prefix, arg))
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

func registerFloat(prefix string, arg *arg.Arg[float64]) {
	verifyArgValue(arg)

	if arg.Required {
		required = append(required, argPath(prefix, arg))
	}

	flagset.Func(argPath(prefix, arg), arg.Desc, floatHandler(prefix, arg))
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

func registerString(prefix string, arg *arg.Arg[string]) {
	verifyArgValue(arg)

	if arg.Required {
		required = append(required, argPath(prefix, arg))
	}

	flagset.Func(argPath(prefix, arg), arg.Desc, stringHandler(prefix, arg))
}

func stringHandler(prefix string, arg *arg.Arg[string]) func(string) error {
	return func(val string) error {
		*arg.Value = val
		return generalHandler(prefix, arg)
	}
}

func registerBool(prefix string, arg *arg.Arg[bool]) {
	verifyArgValue(arg)

	if arg.Required {
		panic(fmt.Errorf("boolean arg was required '%s'", argPath(prefix, arg)))
	}

	flagset.BoolVar(arg.Value, argPath(prefix, arg), *arg.Value, arg.Desc)
}

func verifyArgValue[T any](arg *arg.Arg[T]) {
	if arg.Value == nil {
		panic(fmt.Errorf("Arg.Value must not be a nil pointer, found for for arg '%s'", arg.Name))
	}
}

// Handle required, validation and all other actions.
func generalHandler[T any](prefix string, arg *arg.Arg[T]) error {
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

func argPath[T any](prefix string, arg *arg.Arg[T]) string {
	return fmt.Sprintf("%s.%s", prefix, arg.Name)
}
