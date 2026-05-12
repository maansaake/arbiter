package cli

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/maansaake/arbiter/pkg/module"
	"github.com/spf13/cobra"
)

const FlagsetName = "cli"

var (
	ErrNilPtr       = errors.New("either Arg.Value or Arg.Handler must not be a nil pointer")
	ErrRequiredBool = errors.New("a boolean arg cannot be marked required")
	ErrInvalid      = errors.New("validator failed")
	ErrType         = errors.New("unsupported type")
)

// registerFlags registers all args on cmd's flag set using prefix as a namespace.
// Required flag names are appended to required.
func registerFlags(cmd *cobra.Command, prefix string, args module.Args, required *[]string) error {
	errs := make([]error, 0, len(args))
	for _, arg := range args {
		errs = append(errs, registerFlag(cmd, prefix, arg, required))
	}

	return errors.Join(errs...)
}

// registerFlag dispatches to the type-specific registration function.
func registerFlag(cmd *cobra.Command, prefix string, argument any, required *[]string) error {
	switch a := argument.(type) {
	case *module.Arg[int]:
		return registerIntFlag(cmd, prefix, a, required)
	case *module.Arg[uint]:
		return registerUintFlag(cmd, prefix, a, required)
	case *module.Arg[float64]:
		return registerFloatFlag(cmd, prefix, a, required)
	case *module.Arg[string]:
		return registerStringFlag(cmd, prefix, a, required)
	case *module.Arg[bool]:
		return registerBoolFlag(cmd, prefix, a)
	}

	return ErrType
}

// argFlagValue implements pflag.Value for a typed module.Arg.
// It calls the arg's validator and handler when the flag value is set.
type argFlagValue[T module.TypeConstraint] struct {
	arg      *module.Arg[T]
	parse    func(string) (T, error)
	typeName string
}

func (v *argFlagValue[T]) String() string {
	if v.arg.Value == nil {
		return ""
	}

	return fmt.Sprintf("%v", *v.arg.Value)
}

func (v *argFlagValue[T]) Set(s string) error {
	val, err := v.parse(s)
	if err != nil {
		return err
	}

	*v.arg.Value = val

	if v.arg.Valid != nil && !v.arg.Valid(val) {
		return ErrInvalid
	}

	if v.arg.Handler != nil {
		v.arg.Handler(val)
	}

	return nil
}

func (v *argFlagValue[T]) Type() string {
	return v.typeName
}

func verifyArgValue[T module.TypeConstraint](arg *module.Arg[T]) error {
	if arg.Handler == nil && arg.Value == nil {
		return fmt.Errorf("%w: '%s'", ErrNilPtr, arg.Name)
	} else if arg.Value == nil {
		// For a Handler-only arg, allocate storage so Set() can write to it.
		arg.Value = new(T)
	}

	return nil
}

func registerIntFlag(cmd *cobra.Command, prefix string, arg *module.Arg[int], required *[]string) error {
	if err := verifyArgValue(arg); err != nil {
		return err
	}

	name := argPath(prefix, arg)
	cmd.Flags().Var(&argFlagValue[int]{
		arg: arg,
		parse: func(s string) (int, error) {
			iv, err := strconv.ParseInt(s, 10, 0)
			return int(iv), err
		},
		typeName: "int",
	}, name, arg.Desc)

	if arg.Required {
		*required = append(*required, name)
	}

	return nil
}

func registerUintFlag(cmd *cobra.Command, prefix string, arg *module.Arg[uint], required *[]string) error {
	if err := verifyArgValue(arg); err != nil {
		return err
	}

	name := argPath(prefix, arg)
	cmd.Flags().Var(&argFlagValue[uint]{
		arg: arg,
		parse: func(s string) (uint, error) {
			iv, err := strconv.ParseUint(s, 10, 0)
			return uint(iv), err
		},
		typeName: "uint",
	}, name, arg.Desc)

	if arg.Required {
		*required = append(*required, name)
	}

	return nil
}

func registerFloatFlag(cmd *cobra.Command, prefix string, arg *module.Arg[float64], required *[]string) error {
	if err := verifyArgValue(arg); err != nil {
		return err
	}

	name := argPath(prefix, arg)
	cmd.Flags().Var(&argFlagValue[float64]{
		arg: arg,
		parse: func(s string) (float64, error) {
			return strconv.ParseFloat(s, 64)
		},
		typeName: "float64",
	}, name, arg.Desc)

	if arg.Required {
		*required = append(*required, name)
	}

	return nil
}

func registerStringFlag(cmd *cobra.Command, prefix string, arg *module.Arg[string], required *[]string) error {
	if err := verifyArgValue(arg); err != nil {
		return err
	}

	name := argPath(prefix, arg)
	cmd.Flags().Var(&argFlagValue[string]{
		arg: arg,
		parse: func(s string) (string, error) {
			return s, nil
		},
		typeName: "string",
	}, name, arg.Desc)

	if arg.Required {
		*required = append(*required, name)
	}

	return nil
}

func registerBoolFlag(cmd *cobra.Command, prefix string, arg *module.Arg[bool]) error {
	if err := verifyArgValue(arg); err != nil {
		return err
	}

	if arg.Required {
		return fmt.Errorf("%w: '%s'", ErrRequiredBool, argPath(prefix, arg))
	}

	cmd.Flags().Var(&argFlagValue[bool]{
		arg:      arg,
		parse:    strconv.ParseBool,
		typeName: "bool",
	}, argPath(prefix, arg), arg.Desc)

	return nil
}

func argPath[T module.TypeConstraint](prefix string, arg *module.Arg[T]) string {
	return fmt.Sprintf("%s.%s", prefix, arg.Name)
}
