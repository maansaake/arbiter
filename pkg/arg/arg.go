package arg

type Arg[T any] struct {
	Name string
	Desc string
	Type
	Value   any
	Default any
	Valid   Validator[T]
}

type Type uint

const (
	TypeStr Type = iota
	TypeInt
	TypeUint
	TypeFloat
	TypeBool
)

type Validator[T any] func(v T) bool

type Args []any

func ParseArgs(args Args) error {
	return nil
}
