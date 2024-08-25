package arguments

type Arg struct {
	Name        string
	Description string
	Type        ArgType
	Default     any
	Value       any
	Validator   ArgValidator
}

type Args []*Arg

type ArgType int

const (
	ArgInt ArgType = iota
	ArgRandInt
	ArgString
)

type ArgValidator func(value any) bool
