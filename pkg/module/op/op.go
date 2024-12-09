package op

import "time"

type (
	Op struct {
		Name     string
		Desc     string
		Disabled bool
		Do
		Rate uint
	}
	Ops    []*Op
	Do     func() (Result, error)
	Result struct {
		Name     string
		Duration time.Duration
	}
)
