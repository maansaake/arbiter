package report

import "tres-bon.se/arbiter/pkg/module/op"

type YAMLReporter struct{}

func NewYAML() Reporter {
	return &YAMLReporter{}
}

func (r *YAMLReporter) Op(name string, res *op.Result, err error) {}
func (r *YAMLReporter) LogErr(name string, err error)             {}
func (r *YAMLReporter) LogTrigger(name, result, value string)     {}
func (r *YAMLReporter) CPUErr(string, error)                      {}
func (r *YAMLReporter) CPUTrigger(string, string, float64)        {}
func (r *YAMLReporter) MetricErr(string, error)                   {}
func (r *YAMLReporter) MetricTrigger(string, string, float64)     {}
func (r *YAMLReporter) RSSErr(string, error)                      {}
func (r *YAMLReporter) RSSTrigger(string, string, uint)           {}
func (r *YAMLReporter) VMSErr(string, error)                      {}
func (r *YAMLReporter) VMSTrigger(string, string, uint)           {}
func (r *YAMLReporter) Finalise()                                 {}
