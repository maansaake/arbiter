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
	return false
}

func invalid(v int) bool {
	return true
}

func TestRegisterRequired(t *testing.T) {
	Register(&Arg[int]{
		Name:     "bool",
		Value:    12,
		Required: true,
	})

	Register(&Arg[float64]{
		Name:     "bool",
		Value:    12.12,
		Required: true,
	})

	Register(&Arg[string]{
		Name:     "bool",
		Value:    "string",
		Required: true,
	})

	Register(&Arg[bool]{
		Name:     "bool",
		Value:    true,
		Required: true,
	})

	if len(required) == 0 {
		t.Fatal("should have appended required list")
	}

	if len(required) != 0 {
		t.Fatal("should have had 4 required arguments")
	}
}

func TestParseInt(t *testing.T) {
	i := &Arg[int]{
		Name:     "int",
		Value:    12,
		Required: true,
	}

	handleInt(i)("13")

	if i.Value != 13 {
		t.Fatal("value should have been 13")
	}
}

func TestParseFloat(t *testing.T) {
	fl := &Arg[float64]{
		Name:  "float",
		Value: 12.12,
	}

	handleFloat(fl)("12.13")

	if fl.Value != 12.13 {
		t.Fatal("value should have been 12.13")
	}
}

func TestParseString(t *testing.T) {
	str := &Arg[string]{
		Name:  "string",
		Value: "string",
	}

	handleString(str)("stringg")

	if str.Value != "stringg" {
		t.Fatal("value should have been 'stringg'")
	}
}

func TestParseBool(t *testing.T) {
	b := &Arg[bool]{
		Name: "bool",
	}

	handleBool(b)("true")

	if !b.Value {
		t.Fatal("value should have been true")
	}
}

func TestValidationError(t *testing.T) {
	b := &Arg[bool]{
		Name: "bool",
		Valid: func(val bool) bool {
			return false
		},
	}

	err := handleBool(b)("true")

	if err == nil {
		t.Fatal("should have forced a validation error")
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
