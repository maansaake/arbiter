package mockreport

import (
	"context"

	"tres-bon.se/arbiter/pkg/module/op"
	"tres-bon.se/arbiter/pkg/report"
)

type ReporterMock struct {
	OpResults []*op.Result
	OpErrors  []error
}

func NewMock() report.Reporter {
	return &ReporterMock{
		OpResults: make([]*op.Result, 0, 1),
		OpErrors:  make([]error, 0, 1),
	}
}

// CPU implements report.Reporter.
func (r *ReporterMock) CPU(module string, value float64) {
	panic("unimplemented")
}

// CPUErr implements report.Reporter.
func (r *ReporterMock) CPUErr(module string, err error) {
	panic("unimplemented")
}

// CPUTrigger implements report.Reporter.
func (r *ReporterMock) CPUTrigger(module string, tr *report.TriggerReport[float64]) {
	panic("unimplemented")
}

// Finalise implements report.Reporter.
func (r *ReporterMock) Finalise() error {
	panic("unimplemented")
}

// LogErr implements report.Reporter.
func (r *ReporterMock) LogErr(module string, err error) {
	panic("unimplemented")
}

// LogTrigger implements report.Reporter.
func (r *ReporterMock) LogTrigger(module string, tr *report.TriggerReport[string]) {
	panic("unimplemented")
}

// MetricErr implements report.Reporter.
func (r *ReporterMock) MetricErr(module string, metric string, err error) {
	panic("unimplemented")
}

// MetricTrigger implements report.Reporter.
func (r *ReporterMock) MetricTrigger(module string, metric string, tr *report.TriggerReport[float64]) {
	panic("unimplemented")
}

// Op implements report.Reporter.
func (r *ReporterMock) Op(module, op string, result *op.Result, err error) {
	if err != nil {
		r.OpErrors = append(r.OpErrors, err)
	} else {
		r.OpResults = append(r.OpResults, result)
	}
}

// RSS implements report.Reporter.
func (r *ReporterMock) RSS(module string, value uint) {
	panic("unimplemented")
}

// RSSErr implements report.Reporter.
func (r *ReporterMock) RSSErr(module string, err error) {
	panic("unimplemented")
}

// RSSTrigger implements report.Reporter.
func (r *ReporterMock) RSSTrigger(module string, tr *report.TriggerReport[uint]) {
	panic("unimplemented")
}

// Start implements report.Reporter.
func (r *ReporterMock) Start(context.Context) {
	panic("unimplemented")
}

// VMS implements report.Reporter.
func (r *ReporterMock) VMS(module string, value uint) {
	panic("unimplemented")
}

// VMSErr implements report.Reporter.
func (r *ReporterMock) VMSErr(module string, err error) {
	panic("unimplemented")
}

// VMSTrigger implements report.Reporter.
func (r *ReporterMock) VMSTrigger(module string, tr *report.TriggerReport[uint]) {
	panic("unimplemented")
}
