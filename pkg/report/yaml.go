package report

import (
	"bytes"
	"context"
	"io"
	"os"
	"time"

	"gopkg.in/yaml.v3"
	"tres-bon.se/arbiter/pkg/module/op"
)

type (
	yamlReporter struct {
		path string
		*report
		synchronizer chan func()
	}
	YAMLOpts struct {
		Path string
	}
)

func NewYAML(opts *YAMLOpts) Reporter {
	reporter := &yamlReporter{
		report: &report{
			Start:   time.Now(),
			Modules: make(map[string]*modules),
		},
		path:         opts.Path,
		synchronizer: make(chan func()),
	}

	return reporter
}

func (r *yamlReporter) Start(ctx context.Context) {
	go func() {
		for {
			select {
			case f := <-r.synchronizer:
				f()
			case <-ctx.Done():
				return
			}
		}
	}()
}

func (r *yamlReporter) Op(module string, res *op.Result, err error) {
	r.synchronizer <- r.handleOp(module, res, err)
}

func (r *yamlReporter) handleOp(module string, res *op.Result, err error) func() {
	return func() {
		_, ok := r.report.Modules[module]
		if !ok {
			r.report.Modules[module] = &modules{
				Operations: make(map[string]*operations),
			}
		}

		_, ok = r.report.Modules[module].Operations[res.Name]
		if !ok {
			r.report.Modules[module].Operations[res.Name] = &operations{
				Timing: &operationTiming{},
				Errors: make([]string, 0),
			}
		}
		r.report.Modules[module].Operations[res.Name].Executions++

		if err != nil {
			r.report.Modules[module].Operations[res.Name].Errors = append(r.report.Modules[module].Operations[res.Name].Errors, err.Error())
		} else {
			if r.report.Modules[module].Operations[res.Name].Timing.count == 0 {
				r.report.Modules[module].Operations[res.Name].Timing.Longest = res.Duration
				r.report.Modules[module].Operations[res.Name].Timing.Shortest = res.Duration
				r.report.Modules[module].Operations[res.Name].Timing.Average = res.Duration
			}
			r.report.Modules[module].Operations[res.Name].Timing.count++

			if res.Duration > r.report.Modules[module].Operations[res.Name].Timing.Longest {
				r.report.Modules[module].Operations[res.Name].Timing.Longest = res.Duration
			}
			if res.Duration < r.report.Modules[module].Operations[res.Name].Timing.Shortest {
				r.report.Modules[module].Operations[res.Name].Timing.Shortest = res.Duration
			}

			r.report.Modules[module].Operations[res.Name].Timing.total += res.Duration
			r.report.Modules[module].Operations[res.Name].Timing.Average = r.report.Modules[module].Operations[res.Name].Timing.total / time.Duration(r.report.Modules[module].Operations[res.Name].Timing.count)
		}
	}
}

func (r *yamlReporter) LogErr(module string, err error)                              {}
func (r *yamlReporter) LogTrigger(module, result, value string)                      {}
func (r *yamlReporter) CPU(module string, value float64)                             {}
func (r *yamlReporter) CPUErr(module string, err error)                              {}
func (r *yamlReporter) CPUTrigger(module string, res string, val float64)            {}
func (r *yamlReporter) RSS(module string, value uint)                                {}
func (r *yamlReporter) RSSErr(module string, err error)                              {}
func (r *yamlReporter) RSSTrigger(module string, res string, val uint)               {}
func (r *yamlReporter) VMS(module string, value uint)                                {}
func (r *yamlReporter) VMSErr(module string, err error)                              {}
func (r *yamlReporter) VMSTrigger(module string, res string, val uint)               {}
func (r *yamlReporter) MetricErr(module, metric string, err error)                   {}
func (r *yamlReporter) MetricTrigger(module, metric string, res string, val float64) {}
func (r *yamlReporter) Finalise() error {
	r.report.End = time.Now()
	r.report.Duration = r.report.End.Sub(r.report.Start)

	file, err := os.Create(r.path)
	if err != nil {
		return err
	}
	defer file.Close()

	bs, err := yaml.Marshal(r.report)
	if err != nil {
		return err
	}

	_, err = io.Copy(file, bytes.NewReader(bs))
	return err
}
