package arg

type (
	Arg[T any] struct {
		Name     string
		Desc     string
		Required bool
		Value    *T
		Valid    func(v T) bool
	}
	Args []any
)
