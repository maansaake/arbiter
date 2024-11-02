package op

import "time"

type Op struct {
	Name     string
	Desc     string
	Disabled bool
	Do
	Rate uint
}

type Do func() (Result, error)

type Result struct {
	Duration         time.Duration
	DurationOverride time.Duration
}

type Ops []*Op
