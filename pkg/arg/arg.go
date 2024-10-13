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

func RegisterOp(ns string, op *op.Op) {
	// Enabled
	registerBool(&Arg[bool]{
		Name:  enabledName(ns, op),
		Desc:  enabledDesc(op),
		Value: true,
		onPresent: func(enabled bool) {
			op.Enabled = true
		},
	})
	// Rate
	registerInt(&Arg[int]{
		Name:  rateName(ns, op),
		Desc:  rateDesc(op),
		Value: 60,
		Valid: ValidatorPositiveInteger,
		onPresent: func(rate int) {
			op.Rate = uint(rate)
		},
	})
}

func enabledName(ns string, op *op.Op) string {
	return fmt.Sprintf("%s.%s.enabled", ns, op.Name)
}

func enabledDesc(op *op.Op) string {
	return fmt.Sprintf("enable %s", op.Name)
}

func rateName(ns string, op *op.Op) string {
	return fmt.Sprintf("%s.%s.rate", ns, op.Name)
}

func rateDesc(op *op.Op) string {
	return fmt.Sprintf("rate of %s per minute", op.Name)
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

		return generalHandler(arg)
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

		return generalHandler(arg)
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
		return generalHandler(arg)
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
