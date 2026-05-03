package reportmock

import (
	"context"

	"github.com/maansaake/arbiter/pkg/module"
	"github.com/maansaake/arbiter/pkg/report"
)

type ReporterMock struct {
	OpResults []*module.Result
	OpErrors  []error
}

var _ report.Reporter = &ReporterMock{}

func NewMock() *ReporterMock {
	return &ReporterMock{
		OpResults: make([]*module.Result, 0, 1),
		OpErrors:  make([]error, 0, 1),
	}
}

// Finalise implements [Reporter].
func (r *ReporterMock) Finalise() error {
	panic("unimplemented")
}

// Op implements report.Reporter.
func (r *ReporterMock) Op(_, _ string, result *module.Result, err error) {
	if err != nil {
		r.OpErrors = append(r.OpErrors, err)
	} else {
		r.OpResults = append(r.OpResults, result)
	}
}

// Start implements report.Reporter.
func (r *ReporterMock) Start(context.Context) {
	panic("unimplemented")
}
