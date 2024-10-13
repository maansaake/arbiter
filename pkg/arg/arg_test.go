package arg

import "testing"

type TestArg struct {
	Type  uint
	Value any
	Valid Validator[any]
}

func val1(v bool) bool {
	return false
}

func val2(v int) bool {
	return true
}

func TestSome(t *testing.T) {
	t.Skip()
}
