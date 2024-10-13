package arg

import (
	"testing"
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

func TestSome(t *testing.T) {
	t.Skip()
	Register(&Arg[int]{
		Name:     "arg",
		Desc:     "desc",
		Required: true,
		Valid:    invalid,
	})

	Parse()
}
