package cli

import (
	"errors"
	"flag"
	"testing"

	"tres-bon.se/arbiter/pkg/module/arg"
)

func intValidator(v int) bool {
	return v == 12
}

func TestRegisterMultiple(t *testing.T) {
	flagset = flag.NewFlagSet("cli", flag.ExitOnError)
	required = []string{}

	err := Register("ns", arg.Args{&arg.Arg[int]{
		Name:     "int",
		Value:    new(int),
		Required: true,
	}})
	if err != nil {
		t.Fatal("no error expected:", err)
	}
}

func TestRegisterRequired(t *testing.T) {
	flagset = flag.NewFlagSet("cli", flag.ExitOnError)
	required = []string{}

	err := register("ns", &arg.Arg[int]{
		Name:     "int",
		Value:    new(int),
		Required: true,
	})
	if err != nil {
		t.Fatal("no error expected:", err)
	}
	err = register("ns", &arg.Arg[float64]{
		Name:     "float",
		Value:    new(float64),
		Required: true,
	})
	if err != nil {
		t.Fatal("no error expected:", err)
	}
	err = register("ns", &arg.Arg[string]{
		Name:     "string",
		Value:    new(string),
		Required: true,
	})
	if err != nil {
		t.Fatal("no error expected:", err)
	}
	err = register("ns", &arg.Arg[bool]{
		Name:  "bool",
		Value: new(bool),
	})
	if err != nil {
		t.Fatal("no error expected:", err)
	}

	if len(required) == 0 {
		t.Fatal("should have appended required list")
	}
	if len(required) != 3 {
		t.Fatal("should have had 3 required arguments")
	}
}

func TestRequiredPresent(t *testing.T) {
	flagset = flag.NewFlagSet(FLAGSET, flag.ExitOnError)
	required = []string{}

	err := register("prefix", &arg.Arg[uint]{
		Name:     "count",
		Value:    new(uint),
		Required: true,
	})
	if err != nil {
		t.Fatal("should have not been an error")
	}
	if len(required) != 1 {
		t.Fatal("should have been 1 required flag")
	}

	err = ParseArgs([]string{"-prefix.count=12"})
	if err != nil {
		t.Fatal("parsing failed:", err)
	}

	if len(required) != 0 {
		t.Fatal("number of required should have been cleared after parsing")
	}
}

func TestRequiredMissing(t *testing.T) {
	flagset = flag.NewFlagSet(FLAGSET, flag.ExitOnError)
	required = []string{}

	err := register("prefix", &arg.Arg[uint]{
		Name:     "count",
		Value:    new(uint),
		Required: true,
	})
	if err != nil {
		t.Fatal("should have not been an error")
	}
	if len(required) != 1 {
		t.Fatal("should have been 1 required flag")
	}

	err = ParseArgs([]string{})
	if err == nil {
		t.Fatal("parsing should have failed")
	}

	if !errors.Is(err, arg.ErrParse) {
		t.Fatal("expected a ErrParseError")
	}
}

func TestRequiredBoolean(t *testing.T) {
	err := register("prefix", &arg.Arg[bool]{
		Name:     "master",
		Value:    new(bool),
		Required: true,
	})
	if err == nil {
		t.Fatal("should have returned an error")
	}
}

func TestFlagParseFailure(t *testing.T) {
	flagset = flag.NewFlagSet(FLAGSET, flag.ContinueOnError)
	required = []string{}

	err := register("prefix", &arg.Arg[uint]{
		Name:     "count",
		Value:    new(uint),
		Required: true,
	})
	if err != nil {
		t.Fatal("should have not been an error")
	}

	err = ParseArgs([]string{"-doesnotexist=12"})
	if err == nil {
		t.Fatal("parsing should have failed")
	}
}

func TestParseInt(t *testing.T) {
	val := 12
	i := &arg.Arg[int]{
		Name:     "int",
		Value:    &val,
		Required: true,
	}

	err := intHandler("prefix", i)("13")
	if err != nil {
		t.Errorf("failed to handle int: %v", err)
	}

	if *i.Value != 13 {
		t.Fatal("value should have been 13")
	}
}

func TestParseFloat(t *testing.T) {
	val := 12.12
	fl := &arg.Arg[float64]{
		Name:  "float",
		Value: &val,
	}

	err := floatHandler("prefix", fl)("12.13")
	if err != nil {
		t.Errorf("failed to handle float: %v", err)
	}

	if *fl.Value != 12.13 {
		t.Fatal("value should have been 12.13")
	}
}

func TestParseString(t *testing.T) {
	val := "string"
	str := &arg.Arg[string]{
		Name:  "string",
		Value: &val,
	}

	err := stringHandler("prefix", str)("stringg")
	if err != nil {
		t.Errorf("failed to handle string: %v", err)
	}

	if *str.Value != "stringg" {
		t.Fatal("value should have been 'stringg'")
	}
}

func TestParseIntFailure(t *testing.T) {
	err := intHandler("prefix", &arg.Arg[int]{})("xyz")
	if err == nil {
		t.Fatal("should not have successfully parsed")
	}
}

func TestParseUIntFailure(t *testing.T) {
	err := uintHandler("prefix", &arg.Arg[uint]{})("xyz")
	if err == nil {
		t.Fatal("should not have successfully parsed")
	}
}

func TestParseFloatFailure(t *testing.T) {
	err := floatHandler("prefix", &arg.Arg[float64]{})("xyz")
	if err == nil {
		t.Fatal("should not have successfully parsed")
	}
}

func TestValidationError(t *testing.T) {
	b := &arg.Arg[int]{
		Name:  "bool",
		Value: new(int),
		Valid: func(val int) bool {
			return val != 12
		},
	}

	err := intHandler("prefix", b)("12")

	if err == nil {
		t.Fatal("should have forced a validation error")
	}
}

func TestValidationOk(t *testing.T) {
	iv := 13
	i := &arg.Arg[int]{
		Name:  "int",
		Value: &iv,
		Valid: intValidator,
	}

	err := intHandler("prefix", i)("12")
	if err != nil {
		t.Fatal("should have not been an error")
	}

	if *i.Value != 12 {
		t.Fatal("should have been 12")
	}
}

func TestPanicRegisterNilPointer(t *testing.T) {
	flagset = flag.NewFlagSet("cli", flag.ExitOnError)
	required = []string{}

	err := register("ns", &arg.Arg[float64]{})
	if err == nil {
		t.Fatal("expected register error")
	}
}

func TestParseArgs(t *testing.T) {
	flagset = flag.NewFlagSet("cli", flag.ExitOnError)
	required = []string{}

	i := &arg.Arg[int]{
		Name:  "intt",
		Desc:  "desc",
		Value: new(int),
	}
	s := &arg.Arg[string]{
		Name:  "stringg",
		Desc:  "desc",
		Value: new(string),
	}
	err := register("ns", i)
	if err != nil {
		t.Fatal("should have not been an error")
	}
	err = register("ns", s)
	if err != nil {
		t.Fatal("should have not been an error")
	}

	err = ParseArgs([]string{"-ns.stringg", "strvalue", "-ns.intt", "12"})
	if err != nil {
		t.Fatal(err)
	}

	if *i.Value != 12 {
		t.Fatal("integer should have been 12")
	}

	if *s.Value != "strvalue" {
		t.Fatal("integer should have been 12")
	}
}
