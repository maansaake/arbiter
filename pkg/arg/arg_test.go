package arg

import (
	"flag"
	"testing"

	"tres-bon.se/arbiter/pkg/module/op"
)

type TestArg struct {
	Type  uint
	Value any
	Valid Validator[any]
}

func validInt(v int) bool {
	return v == 12
}

func TestRegisterRequired(t *testing.T) {
	flagset = flag.NewFlagSet("cli", flag.ExitOnError)
	required = []string{}

	Register("ns", &Arg[int]{
		Name:     "int",
		Value:    new(int),
		Required: true,
	})

	Register("ns", &Arg[float64]{
		Name:     "float",
		Value:    new(float64),
		Required: true,
	})

	Register("ns", &Arg[string]{
		Name:     "string",
		Value:    new(string),
		Required: true,
	})

	Register("ns", &Arg[bool]{
		Name:  "bool",
		Value: new(bool),
	})

	if len(required) == 0 {
		t.Fatal("should have appended required list")
	}

	if len(required) != 3 {
		t.Fatal("should have had 3 required arguments")
	}
}

func TestParseInt(t *testing.T) {
	val := 12
	i := &Arg[int]{
		Name:     "int",
		Value:    &val,
		Required: true,
	}

	err := handleInt(i)("13")
	if err != nil {
		t.Errorf("failed to handle int: %v", err)
	}

	if *i.Value != 13 {
		t.Fatal("value should have been 13")
	}
}

func TestParseFloat(t *testing.T) {
	val := 12.12
	fl := &Arg[float64]{
		Name:  "float",
		Value: &val,
	}

	err := handleFloat(fl)("12.13")
	if err != nil {
		t.Errorf("failed to handle float: %v", err)
	}

	if *fl.Value != 12.13 {
		t.Fatal("value should have been 12.13")
	}
}

func TestParseString(t *testing.T) {
	val := "string"
	str := &Arg[string]{
		Name:  "string",
		Value: &val,
	}

	err := handleString(str)("stringg")
	if err != nil {
		t.Errorf("failed to handle string: %v", err)
	}

	if *str.Value != "stringg" {
		t.Fatal("value should have been 'stringg'")
	}
}

func TestValidationError(t *testing.T) {
	b := &Arg[int]{
		Name:  "bool",
		Value: new(int),
		Valid: func(val int) bool {
			return val != 12
		},
	}

	err := handleInt(b)("12")

	if err == nil {
		t.Fatal("should have forced a validation error")
	}
}

func TestValidationOk(t *testing.T) {
	iv := 13
	i := &Arg[int]{
		Name:  "int",
		Value: &iv,
		Valid: validInt,
	}

	err := handleInt(i)("12")
	if err != nil {
		t.Fatal("should have not been an error")
	}

	if *i.Value != 12 {
		t.Fatal("should have been 12")
	}
}

func TestPanicUnrecognizedType(t *testing.T) {
	flagset = flag.NewFlagSet("cli", flag.ExitOnError)
	required = []string{}
	defer func() {
		t.Log("panicked?")
		if err := recover(); err == nil {
			t.Fatal("expected to recover from panic")
		}
	}()
	Register("ns", &Arg[float32]{})
	t.Fatal("no panic? :(")
}

func TestPanicRegisterNilPointer(t *testing.T) {
	flagset = flag.NewFlagSet("cli", flag.ExitOnError)
	required = []string{}
	defer func() {
		t.Log("panicked?")
		if err := recover(); err == nil {
			t.Fatal("expected to recover from panic")
		}
	}()
	Register("ns", &Arg[float64]{})
	t.Fatal("no panic? :(")
}

func TestParse(t *testing.T) {
	flagset = flag.NewFlagSet("cli", flag.ExitOnError)
	required = []string{}
	i := &Arg[int]{
		Name:  "intt",
		Desc:  "desc",
		Value: new(int),
	}
	s := &Arg[string]{
		Name:  "stringg",
		Desc:  "desc",
		Value: new(string),
	}
	Register("ns", i)
	Register("ns", s)
	op := &op.Op{
		Name: "op",
		Desc: "desc",
	}
	RegisterOp("ns", op)

	Parse([]string{"-ns.stringg", "strvalue", "-ns.intt", "12", "-ns.op.disable", "-ns.op.rate", "12"})

	if *i.Value != 12 {
		t.Fatal("integer should have been 12")
	}

	if *s.Value != "strvalue" {
		t.Fatal("integer should have been 12")
	}

	if !op.Disabled {
		t.Fatal("should have been disabled")
	}

	if op.Rate != 12 {
		t.Fatal("rate should have been 12")
	}
}
