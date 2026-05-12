package cli

import (
	"errors"
	"fmt"
	"strconv"
	"testing"

	"github.com/maansaake/arbiter/pkg/module"
	"github.com/spf13/cobra"
)

func intValidator(v int) bool {
	return v == 12
}

func newTestCmd() *cobra.Command {
	return &cobra.Command{Use: "test", SilenceErrors: true, SilenceUsage: true}
}

func TestRegisterMultiple(t *testing.T) {
	cmd := newTestCmd()
	var required []string

	err := registerFlags(cmd, "ns", module.Args{&module.Arg[int]{
		Name:     "int",
		Value:    new(int),
		Required: true,
	}}, &required)
	if err != nil {
		t.Fatal("no error expected:", err)
	}
}

func TestRegisterRequired(t *testing.T) {
	cmd := newTestCmd()
	var required []string

	err := registerFlag(cmd, "ns", &module.Arg[int]{Name: "int", Value: new(int), Required: true}, &required)
	if err != nil {
		t.Fatal("no error expected:", err)
	}

	err = registerFlag(cmd, "ns", &module.Arg[float64]{Name: "float", Value: new(float64), Required: true}, &required)
	if err != nil {
		t.Fatal("no error expected:", err)
	}

	err = registerFlag(cmd, "ns", &module.Arg[string]{Name: "string", Value: new(string), Required: true}, &required)
	if err != nil {
		t.Fatal("no error expected:", err)
	}

	err = registerFlag(cmd, "ns", &module.Arg[bool]{Name: "bool", Value: new(bool)}, &required)
	if err != nil {
		t.Fatal("no error expected:", err)
	}

	if len(required) == 0 {
		t.Fatal("should have appended required list")
	}

	if len(required) != 3 {
		t.Fatalf("should have had 3 required arguments, got %d", len(required))
	}
}

func TestRequiredPresent(t *testing.T) {
	cmd := newTestCmd()
	var required []string

	err := registerFlag(cmd, "prefix", &module.Arg[uint]{Name: "count", Value: new(uint), Required: true}, &required)
	if err != nil {
		t.Fatal("should have not been an error")
	}

	if len(required) != 1 {
		t.Fatal("should have been 1 required flag")
	}

	err = cmd.ParseFlags([]string{"--prefix.count=12"})
	if err != nil {
		t.Fatal("parsing failed:", err)
	}

	for _, name := range required {
		if !cmd.Flags().Changed(name) {
			t.Fatalf("flag %s should have been marked as changed", name)
		}
	}
}

func TestRequiredMissing(t *testing.T) {
	cmd := newTestCmd()
	var required []string

	err := registerFlag(cmd, "prefix", &module.Arg[uint]{Name: "count", Value: new(uint), Required: true}, &required)
	if err != nil {
		t.Fatal("should have not been an error")
	}

	if len(required) != 1 {
		t.Fatal("should have been 1 required flag")
	}

	// Simulate the required check that happens in cmd.RunE.
	var missingErr error

	for _, name := range required {
		if !cmd.Flags().Changed(name) {
			missingErr = errors.Join(missingErr, fmt.Errorf("%w: --%s is required", module.ErrArgRequired, name))
		}
	}

	if missingErr == nil {
		t.Fatal("parsing should have failed")
	}

	if !errors.Is(missingErr, module.ErrArgRequired) {
		t.Fatal("expected a ErrArgRequired, got", missingErr)
	}
}

func TestRequiredBoolean(t *testing.T) {
	cmd := newTestCmd()
	var required []string

	err := registerFlag(cmd, "prefix", &module.Arg[bool]{
		Name:     "master",
		Value:    new(bool),
		Required: true,
	}, &required)
	if err == nil {
		t.Fatal("should have returned an error")
	}

	if !errors.Is(err, ErrRequiredBool) {
		t.Fatalf("expected ErrRequiredBool, got %v", err)
	}
}

func TestFlagParseFailure(t *testing.T) {
	cmd := newTestCmd()
	var required []string

	err := registerFlag(cmd, "prefix", &module.Arg[uint]{Name: "count", Value: new(uint), Required: true}, &required)
	if err != nil {
		t.Fatal("should have not been an error")
	}

	err = cmd.ParseFlags([]string{"--doesnotexist=12"})
	if err == nil {
		t.Fatal("parsing should have failed")
	}
}

func TestParseInt(t *testing.T) {
	val := 12
	i := &module.Arg[int]{
		Name:     "int",
		Value:    &val,
		Required: true,
	}

	v := &argFlagValue[int]{
		arg: i,
		parse: func(s string) (int, error) {
			iv, err := strconv.ParseInt(s, 10, 0)
			return int(iv), err
		},
	}

	if err := v.Set("13"); err != nil {
		t.Errorf("failed to handle int: %v", err)
	}

	if *i.Value != 13 {
		t.Fatal("value should have been 13")
	}
}

func TestParseFloat(t *testing.T) {
	val := 12.12
	fl := &module.Arg[float64]{
		Name:  "float",
		Value: &val,
	}

	v := &argFlagValue[float64]{
		arg: fl,
		parse: func(s string) (float64, error) {
			return strconv.ParseFloat(s, 64)
		},
	}

	if err := v.Set("12.13"); err != nil {
		t.Errorf("failed to handle float: %v", err)
	}

	if *fl.Value != 12.13 {
		t.Fatal("value should have been 12.13")
	}
}

func TestParseString(t *testing.T) {
	val := "string"
	str := &module.Arg[string]{
		Name:  "string",
		Value: &val,
	}

	v := &argFlagValue[string]{
		arg:   str,
		parse: func(s string) (string, error) { return s, nil },
	}

	if err := v.Set("stringg"); err != nil {
		t.Errorf("failed to handle string: %v", err)
	}

	if *str.Value != "stringg" {
		t.Fatal("value should have been 'stringg'")
	}
}

func TestParseIntFailure(t *testing.T) {
	v := &argFlagValue[int]{
		arg: &module.Arg[int]{Value: new(int)},
		parse: func(s string) (int, error) {
			iv, err := strconv.ParseInt(s, 10, 0)
			return int(iv), err
		},
	}

	if err := v.Set("xyz"); err == nil {
		t.Fatal("should not have successfully parsed")
	}
}

func TestParseUIntFailure(t *testing.T) {
	v := &argFlagValue[uint]{
		arg: &module.Arg[uint]{Value: new(uint)},
		parse: func(s string) (uint, error) {
			iv, err := strconv.ParseUint(s, 10, 0)
			return uint(iv), err
		},
	}

	if err := v.Set("xyz"); err == nil {
		t.Fatal("should not have successfully parsed")
	}
}

func TestParseFloatFailure(t *testing.T) {
	v := &argFlagValue[float64]{
		arg: &module.Arg[float64]{Value: new(float64)},
		parse: func(s string) (float64, error) {
			return strconv.ParseFloat(s, 64)
		},
	}

	if err := v.Set("xyz"); err == nil {
		t.Fatal("should not have successfully parsed")
	}
}

func TestValidationError(t *testing.T) {
	v := &argFlagValue[int]{
		arg: &module.Arg[int]{
			Name:  "bool",
			Value: new(int),
			Valid: func(val int) bool { return val != 12 },
		},
		parse: func(s string) (int, error) {
			iv, err := strconv.ParseInt(s, 10, 0)
			return int(iv), err
		},
	}

	if err := v.Set("12"); err == nil {
		t.Fatal("should have forced a validation error")
	}
}

func TestValidationOk(t *testing.T) {
	iv := 13
	v := &argFlagValue[int]{
		arg: &module.Arg[int]{
			Name:  "int",
			Value: &iv,
			Valid: intValidator,
		},
		parse: func(s string) (int, error) {
			parsed, err := strconv.ParseInt(s, 10, 0)
			return int(parsed), err
		},
	}

	if err := v.Set("12"); err != nil {
		t.Fatal("should have not been an error")
	}

	if iv != 12 {
		t.Fatal("should have been 12")
	}
}

func TestPanicRegisterNilPointer(t *testing.T) {
	cmd := newTestCmd()
	var required []string

	err := registerFlag(cmd, "ns", &module.Arg[float64]{}, &required)
	if err == nil {
		t.Fatal("expected register error")
	}
}

func TestParseArgs(t *testing.T) {
	cmd := newTestCmd()
	var required []string

	i := &module.Arg[int]{Name: "intt", Desc: "desc", Value: new(int)}
	s := &module.Arg[string]{Name: "stringg", Desc: "desc", Value: new(string)}

	if err := registerFlag(cmd, "ns", i, &required); err != nil {
		t.Fatal("should have not been an error:", err)
	}

	if err := registerFlag(cmd, "ns", s, &required); err != nil {
		t.Fatal("should have not been an error:", err)
	}

	if err := cmd.ParseFlags([]string{"--ns.stringg", "strvalue", "--ns.intt", "12"}); err != nil {
		t.Fatal(err)
	}

	if *i.Value != 12 {
		t.Fatal("integer should have been 12")
	}

	if *s.Value != "strvalue" {
		t.Fatal("string should have been 'strvalue'")
	}
}
