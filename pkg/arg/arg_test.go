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

func valid(v bool) bool {
	return v
}

func validInt(v int) bool {
	return v == 12
}

func TestRegisterRequired(t *testing.T) {
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
		Name:     "bool",
		Value:    new(bool),
		Required: true,
	})

	if len(required) == 0 {
		t.Fatal("should have appended required list")
	}

	if len(required) != 4 {
		t.Fatal("should have had 4 required arguments")
	}
}

func TestParseInt(t *testing.T) {
	val := 12
	i := &Arg[int]{
		Name:     "int",
		Value:    &val,
		Required: true,
	}

	handleInt(i)("13")

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

	handleFloat(fl)("12.13")

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

	handleString(str)("stringg")

	if *str.Value != "stringg" {
		t.Fatal("value should have been 'stringg'")
	}
}

func TestParseBool(t *testing.T) {
	b := &Arg[bool]{
		Name:  "bool",
		Value: new(bool),
	}

	handleBool(b)("true")

	if !*b.Value {
		t.Fatal("value should have been true")
	}
}

func TestValidationError(t *testing.T) {
	b := &Arg[bool]{
		Name:  "bool",
		Value: new(bool),
		Valid: func(val bool) bool {
			return false
		},
	}

	err := handleBool(b)("true")

	if err == nil {
		t.Fatal("should have forced a validation error")
	}
}

func TestValidationOk(t *testing.T) {
	b := &Arg[bool]{
		Name:  "bool",
		Value: new(bool),
		Valid: valid,
	}
	iv := 13
	i := &Arg[int]{
		Name:  "int",
		Value: &iv,
		Valid: validInt,
	}

	err := handleBool(b)("false")

	if err == nil {
		t.Fatal("should have forced a validation error")
	}

	err = handleInt(i)("12")
	if err != nil {
		t.Fatal("should have not been an error")
	}

	if *i.Value != 12 {
		t.Fatal("should have been 12")
	}
}

func TestRegisterOp(t *testing.T) {
	op := &op.Op{
		Name: "op",
		Desc: "this is op",
	}

	RegisterOp("ns", op)

	t.Log(flag.CommandLine)
}
