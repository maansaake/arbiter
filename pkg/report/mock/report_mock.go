package mockreport

import (
	"tres-bon.se/arbiter/pkg/module/op"
	"tres-bon.se/arbiter/pkg/report"
)

type ReporterMock struct {
	Results []*op.Result
	Errors  []error
}

func NewReporterMock() report.Reporter {
	return &ReporterMock{
		Results: make([]*op.Result, 0, 1),
		Errors:  make([]error, 0, 1),
	}
}

// Finalise implements report.Reporter.
func (r *ReporterMock) Finalise() {
	panic("unimplemented")
}

// Op implements report.Reporter.
func (r *ReporterMock) Op(res *op.Result, err error) {
	r.Results = append(r.Results, res)
	r.Errors = append(r.Errors, err)
}
