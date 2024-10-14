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

	onPresent func(val T)
}

type Args []any

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
	// Enabled
	registerBool(ns, &Arg[bool]{
		Name:  enabledName(op),
		Desc:  enabledDesc(op),
		Value: true,
		onPresent: func(enabled bool) {
			op.Enabled = true
		},
	})
	// Rate
	registerInt(ns, &Arg[int]{
		Name:  rateName(op),
		Desc:  rateDesc(op),
		Value: 60,
		Valid: ValidatorPositiveInteger,
		onPresent: func(rate int) {
			op.Rate = uint(rate)
		},
	})
}

func enabledName(op *op.Op) string {
	return fmt.Sprintf("%s.enabled", op.Name)
}

func enabledDesc(op *op.Op) string {
	return fmt.Sprintf("enable %s", op.Name)
}

func rateName(op *op.Op) string {
	return fmt.Sprintf("%s.rate", op.Name)
}

func rateDesc(op *op.Op) string {
	return fmt.Sprintf("rate of %s per minute", op.Name)
}

func Parse(args []string) {
	flagset.Parse(args)

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

func registerInt(ns string, arg *Arg[int]) {
	if arg.Required {
		required = append(required, arg.Name)
	}

	flagset.Func(fmt.Sprintf("%s.%s", ns, arg.Name), arg.Desc, handleInt(arg))
}

func handleInt(arg *Arg[int]) func(string) error {
	return func(val string) error {
		iv, err := strconv.ParseInt(val, 10, 0)
		if err != nil {
			return err
		}

		arg.Value = int(iv)

		return generalHandler(arg)
	}
}

func registerFloat(ns string, arg *Arg[float64]) {
	if arg.Required {
		required = append(required, arg.Name)
	}

	flagset.Func(fmt.Sprintf("%s.%s", ns, arg.Name), arg.Desc, handleFloat(arg))
}

func handleFloat(arg *Arg[float64]) func(string) error {
	return func(val string) error {
		fv, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return err
		}

		arg.Value = fv

		return generalHandler(arg)
	}
}

func registerString(ns string, arg *Arg[string]) {
	if arg.Required {
		required = append(required, arg.Name)
	}

	flagset.Func(fmt.Sprintf("%s.%s", ns, arg.Name), arg.Desc, handleString(arg))
}

func handleString(arg *Arg[string]) func(string) error {
	return func(val string) error {
		arg.Value = val
		return generalHandler(arg)
	}
}

func registerBool(ns string, arg *Arg[bool]) {
	if arg.Required {
		required = append(required, arg.Name)
	}

	flagset.Func(fmt.Sprintf("%s.%s", ns, arg.Name), arg.Desc, handleBool(arg))
}

func handleBool(arg *Arg[bool]) func(string) error {
	return func(val string) error {
		b, err := strconv.ParseBool(val)
		if err != nil {
			return err
		}

		arg.Value = b

		return generalHandler(arg)
	}
}

// Handle required, validation and all other actions.
func generalHandler[T any](arg *Arg[T]) error {
	if arg.Required {
		// Find and pop arg from required slice
		for i, an := range required {
			if an == arg.Name {
				required = slices.Delete(required, i, i+1)
			}
		}
	}

	if arg.Valid != nil && !arg.Valid(arg.Value) {
		return fmt.Errorf("%w: argument '%s' has invalid value '%v'", ErrInvalid, arg.Name, arg.Value)
	}

	// If operation register, run onPresent
	if arg.onPresent != nil {
		arg.onPresent(arg.Value)
	}

	return nil
}
