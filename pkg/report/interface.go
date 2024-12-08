package report

import "tres-bon.se/arbiter/pkg/module/op"

type Reporter interface {
	Op(string, *op.Result, error)
	LogErr(string, error)
	LogTrigger(string, string, string)
	CPUErr(string, error)
	CPUTrigger(string, string, float64)
	RSSErr(string, error)
	RSSTrigger(string, string, uint)
	VMSErr(string, error)
	VMSTrigger(string, string, uint)
	MetricErr(string, error)
	MetricTrigger(string, string, float64)
	Finalise()
}
