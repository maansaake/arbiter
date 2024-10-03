package arg

type Arg struct {
	Name string
	Desc string
	Type
	Value   any
	Default any
	Validator[any]
}

type Type uint

const (
	TypeStr Type = iota
	TypeInt
	TypeUint
	TypeFloat
	TypeBool
)

type Validator[T any] func(T) bool

type Args []*Arg

func ParseArgs(args Args) error {
	return nil
}
