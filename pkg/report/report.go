package report

import "tres-bon.se/arbiter/pkg/module/op"

type Reporter interface {
	Op(*op.Result, error)
	Finalise()
}
