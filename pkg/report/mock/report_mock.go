package mockreport

import (
	"context"

	"tres-bon.se/arbiter/pkg/module/op"
	"tres-bon.se/arbiter/pkg/report"
)

type ReporterMock struct {
	Results []*op.Result
	Errors  []error
}

// Start implements report.Reporter.
func (r *ReporterMock) Start(context.Context) {
	panic("unimplemented")
}

func NewReporterMock() report.Reporter {
	return &ReporterMock{
		Results: make([]*op.Result, 0, 1),
		Errors:  make([]error, 0, 1),
	}
}

// Finalise implements report.Reporter.
func (r *ReporterMock) Finalise() error {
	panic("unimplemented")
}

// CPU implements report.Reporter.
func (r *ReporterMock) CPU(module string, value float64) {
	panic("unimplemented")
}

// RSS implements report.Reporter.
func (r *ReporterMock) RSS(module string, value uint) {
	panic("unimplemented")
}

// VMS implements report.Reporter.
func (r *ReporterMock) VMS(module string, value uint) {
	panic("unimplemented")
}

// Op implements report.Reporter.
func (r *ReporterMock) Op(name string, res *op.Result, err error) {
	r.Results = append(r.Results, res)
	r.Errors = append(r.Errors, err)
}

func (r *ReporterMock) LogErr(name string, err error) {}

// LogTrigger implements report.Reporter.
func (r *ReporterMock) LogTrigger(string, string, string) {
	panic("unimplemented")
}

// CPUErr implements report.Reporter.
func (r *ReporterMock) CPUErr(string, error) {
	panic("unimplemented")
}

// CPUTrigger implements report.Reporter.
func (r *ReporterMock) CPUTrigger(string, string, float64) {
	panic("unimplemented")
}

// MetricErr implements report.Reporter.
func (r *ReporterMock) MetricErr(string, string, error) {
	panic("unimplemented")
}

// MetricTrigger implements report.Reporter.
func (r *ReporterMock) MetricTrigger(string, string, string, float64) {
	panic("unimplemented")
}

// RSSErr implements report.Reporter.
func (r *ReporterMock) RSSErr(string, error) {
	panic("unimplemented")
}

// RSSTrigger implements report.Reporter.
func (r *ReporterMock) RSSTrigger(string, string, uint) {
	panic("unimplemented")
}

// VMSErr implements report.Reporter.
func (r *ReporterMock) VMSErr(string, error) {
	panic("unimplemented")
}

// VMSTrigger implements report.Reporter.
func (r *ReporterMock) VMSTrigger(string, string, uint) {
	panic("unimplemented")
}
