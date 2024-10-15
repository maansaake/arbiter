package arg

type Validator[T any] func(v T) bool
