package report

import "tres-bon.se/arbiter/pkg/module/op"

type YAMLReporter struct{}

func NewYAML() Reporter {
	return &YAMLReporter{}
}

func (r *YAMLReporter) Op(res *op.Result, err error) {}

func (r *YAMLReporter) Finalise() {}
