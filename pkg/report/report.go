package report

import (
	"context"

	"github.com/maansaake/arbiter/pkg/module"
)

type (
	Reporter interface {
		Start(context.Context)
		ReportError(err error)
		ReportOp(module, op string, result *module.Result, err error)
		Finalise() error
	}
)
