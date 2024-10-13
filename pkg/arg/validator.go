package arg

type Validator[T any] func(v T) bool

func ValidatorPositiveInteger(val int) bool {
	return val > -1
}
