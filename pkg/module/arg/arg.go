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
	Value    *T
	Valid    func(v T) bool
	ns       string
}

type Args []any

func (a *Arg[T]) argPath() string {
	return fmt.Sprintf("%s.%s", a.ns, a.Name)
}

var (
	ErrInvalid = errors.New("invalid value")

	flagset  = flag.NewFlagSet(FLAGSET, flag.ExitOnError)
	required []string
)

const (
	FLAGSET = "cli"
)

func Register(ns string, arg any) {
	switch ta := arg.(type) {
	case *Arg[int]:
		registerInt(ns, ta)
	case *Arg[float64]:
		registerFloat(ns, ta)
	case *Arg[string]:
		registerString(ns, ta)
	case *Arg[bool]:
		registerBool(ns, ta)
	default:
		panic("not implemented")
	}
}

func RegisterOp(ns string, op *op.Op) {
	// Disable
	registerBool(ns, &Arg[bool]{
		Name:  disableName(op),
		Desc:  disableDesc(op),
		Value: &op.Disabled,
	})
	// Rate
	registerUint(ns, &Arg[uint]{
		Name:  rateName(op),
		Desc:  rateDesc(op),
		Value: &op.Rate,
	})
}

func disableName(op *op.Op) string {
	return fmt.Sprintf("%s.disable", op.Name)
}

func disableDesc(op *op.Op) string {
	return fmt.Sprintf("disable %s", op.Name)
}

func rateName(op *op.Op) string {
	return fmt.Sprintf("%s.rate", op.Name)
}

func rateDesc(op *op.Op) string {
	return fmt.Sprintf("rate of %s per minute", op.Name)
}

func Parse(args []string) {
	err := flagset.Parse(args)
	if err != nil {
		flagset.Usage()
		os.Exit(1)
	}

	// Print all problems before exiting.
	problem := false

	if len(required) > 0 {
		problem = true
		for _, an := range required {
			fmt.Fprintf(os.Stderr, "missing required argument -%s\n", an)
		}
	}

	if problem {
		flagset.SetOutput(os.Stderr)
		flagset.Usage()
		os.Exit(1)
	}
}

func registerInt(ns string, arg *Arg[int]) {
	verifyArgValue(arg)

	arg.ns = ns

	if arg.Required {
		required = append(required, arg.argPath())
	}

	flagset.Func(arg.argPath(), arg.Desc, handleInt(arg))
}

func handleInt(arg *Arg[int]) func(string) error {
	return func(val string) error {
		iv, err := strconv.ParseInt(val, 10, 0)
		if err != nil {
			return err
		}
		*arg.Value = int(iv)

		return generalHandler(arg)
	}
}

func registerUint(ns string, arg *Arg[uint]) {
	verifyArgValue(arg)

	arg.ns = ns

	if arg.Required {
		required = append(required, arg.argPath())
	}

	flagset.Func(arg.argPath(), arg.Desc, handleUint(arg))
}

func handleUint(arg *Arg[uint]) func(string) error {
	return func(val string) error {
		iv, err := strconv.ParseUint(val, 10, 0)
		if err != nil {
			return err
		}
		*arg.Value = uint(iv)

		return generalHandler(arg)
	}
}

func registerFloat(ns string, arg *Arg[float64]) {
	verifyArgValue(arg)

	arg.ns = ns

	if arg.Required {
		required = append(required, arg.argPath())
	}

	flagset.Func(arg.argPath(), arg.Desc, handleFloat(arg))
}

func handleFloat(arg *Arg[float64]) func(string) error {
	return func(val string) error {
		fv, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return err
		}
		*arg.Value = fv

		return generalHandler(arg)
	}
}

func registerString(ns string, arg *Arg[string]) {
	verifyArgValue(arg)

	arg.ns = ns

	if arg.Required {
		required = append(required, arg.argPath())
	}

	flagset.Func(arg.argPath(), arg.Desc, handleString(arg))
}

func handleString(arg *Arg[string]) func(string) error {
	return func(val string) error {
		*arg.Value = val
		return generalHandler(arg)
	}
}

func registerBool(ns string, arg *Arg[bool]) {
	verifyArgValue(arg)

	arg.ns = ns

	if arg.Required {
		panic(fmt.Errorf("boolean arg was required '%s'", arg.argPath()))
	}

	flagset.BoolVar(arg.Value, arg.argPath(), *arg.Value, arg.Desc)
}

func verifyArgValue[T any](arg *Arg[T]) {
	if arg.Value == nil {
		panic(fmt.Errorf("Arg.Value must not be a nil pointer, found for for arg '%s'", arg.Name))
	}
}

// Handle required, validation and all other actions.
func generalHandler[T any](arg *Arg[T]) error {
	if arg.Required {
		// Find and pop arg from required slice
		for i, an := range required {
			if an == arg.argPath() {
				required = slices.Delete(required, i, i+1)
			}
		}
	}

	if arg.Valid != nil && !arg.Valid(*arg.Value) {
		return fmt.Errorf("%w: argument '%s' has invalid value '%v'", ErrInvalid, arg.argPath(), arg.Value)
	}

	return nil
}
