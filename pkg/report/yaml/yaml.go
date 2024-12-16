package yaml

import (
	"bytes"
	"context"
	"io"
	"os"
	"time"

	"gopkg.in/yaml.v3"
	"tres-bon.se/arbiter/pkg/module/op"
	"tres-bon.se/arbiter/pkg/report"
)

type (
	yamlReporter struct {
		path string
		*yamlReport
		synchronizer chan func()
	}
	Opts struct {
		Path string
	}
	yamlReport struct {
		Start    time.Time          `yaml:"start"`
		End      time.Time          `yaml:"end"`
		Duration time.Duration      `yaml:"duration"`
		Modules  map[string]*module `yaml:"modules"`
	}
)

func New(opts *Opts) report.Reporter {
	reporter := &yamlReporter{
		yamlReport: &yamlReport{
			Start:   time.Now(),
			Modules: make(map[string]*module),
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
			case <-ctx.Done():
				return
			case f := <-r.synchronizer:
				f()
			}
		}
	}()
}

func (r *yamlReporter) Op(mod string, res *op.Result, err error) {
	r.synchronizer <- func() {
		r.yamlReport.module(mod).addOp(res, err)
	}
}

func (r *yamlReporter) LogErr(mod string, err error) {
	r.synchronizer <- func() {
		r.yamlReport.module(mod).addLogErr(err)
	}
}

func (r *yamlReporter) LogTrigger(mod string, tr *report.TriggerReport[string]) {
	r.synchronizer <- func() {
		r.yamlReport.module(mod).addLogTrigger(tr)
	}
}

func (r *yamlReporter) CPU(mod string, value float64) {
	r.synchronizer <- func() {
		r.yamlReport.module(mod).addCPU(value)
	}
}

func (r *yamlReporter) CPUErr(mod string, err error) {
	r.synchronizer <- func() {
		r.yamlReport.module(mod).addCPUErr(err)
	}
}

func (r *yamlReporter) CPUTrigger(mod string, tr *report.TriggerReport[float64]) {
	r.synchronizer <- func() {
		r.yamlReport.module(mod).addCPUTrigger(tr)
	}
}

func (r *yamlReporter) RSS(mod string, value uint) {
	r.synchronizer <- func() {
		r.yamlReport.module(mod).addRSS(value)
	}
}

func (r *yamlReporter) RSSErr(mod string, err error) {
	r.synchronizer <- func() {
		r.yamlReport.module(mod).addRSSErr(err)
	}
}

func (r *yamlReporter) RSSTrigger(mod string, tr *report.TriggerReport[uint]) {
	r.synchronizer <- func() {
		r.yamlReport.module(mod).addRSSTrigger(tr)
	}
}

func (r *yamlReporter) VMS(mod string, value uint) {
	r.synchronizer <- func() {
		r.yamlReport.module(mod).addVMS(value)
	}
}

func (r *yamlReporter) VMSErr(mod string, err error) {
	r.synchronizer <- func() {
		r.yamlReport.module(mod).addVMSErr(err)
	}
}

func (r *yamlReporter) VMSTrigger(mod string, tr *report.TriggerReport[uint]) {
	r.synchronizer <- func() {
		r.yamlReport.module(mod).addVMSTrigger(tr)
	}
}

func (r *yamlReporter) MetricErr(mod, metric string, err error) {
	r.synchronizer <- func() {
		r.yamlReport.module(mod).addMetricErr(err)
	}
}

func (r *yamlReporter) MetricTrigger(mod, metric string, tr *report.TriggerReport[float64]) {
	r.synchronizer <- func() {
		r.yamlReport.module(mod).addMetricTrigger(tr)
	}
}

func (r *yamlReporter) Finalise() error {
	r.yamlReport.End = time.Now()
	r.yamlReport.Duration = r.yamlReport.End.Sub(r.yamlReport.Start)

	file, err := os.Create(r.path)
	if err != nil {
		return err
	}
	defer file.Close()

	bs, err := yaml.Marshal(r.yamlReport)
	if err != nil {
		return err
	}

	_, err = io.Copy(file, bytes.NewReader(bs))
	return err
}

/*INTERNAL*/

func (r *yamlReport) module(mod string) (m *module) {
	m, ok := r.Modules[mod]
	if !ok {
		m = newModule()
		r.Modules[mod] = m
	}

	return m
}
