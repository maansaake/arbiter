package yaml

import (
	"context"
	"os"
	"time"

	"gopkg.in/yaml.v3"
	"tres-bon.se/arbiter/pkg/module/op"
	"tres-bon.se/arbiter/pkg/report"
)

type (
	// Options for the YAML reporter.
	Opts struct {
		// Start time to set in the report. If left empty a start time is set
		// when calling `New()`.
		Start time.Time
		// The final path of the YAML report.
		Path string
		// The buffer size sets the number of buffered report calls that are yet
		// to be handled. Values < 1 will be ignored.
		Buffer int
	}
	// Implements the reporter interface.
	YAMLReporter struct {
		// The final path of the YAML report.
		path string
		// The YAML report.
		report *yamlReport
		// Synchronizer channel to limit access to the report to 1 thread. Also
		// speeds up calls to the reporter interface.
		synchronizer chan func()
		stopped      chan bool
	}
	// Struct for the YAML report. The entire thing is marshaled into the final
	// file.
	yamlReport struct {
		Start    time.Time          `yaml:"start"`
		End      time.Time          `yaml:"end"`
		Duration time.Duration      `yaml:"duration"`
		Modules  map[string]*module `yaml:"modules"`
	}
)

var _ report.Reporter = &YAMLReporter{}

// Create a new YAML reporter.
func New(opts *Opts) *YAMLReporter {
	var start time.Time
	var buffer int
	if opts.Buffer > 0 {
		buffer = opts.Buffer
	} else {
		buffer = 100
	}

	if opts.Start.IsZero() {
		start = time.Now()
	} else {
		start = opts.Start
	}

	reporter := &YAMLReporter{
		report: &yamlReport{
			Start:   start,
			Modules: make(map[string]*module),
		},
		path:         opts.Path,
		synchronizer: make(chan func(), buffer),
		stopped:      make(chan bool),
	}

	return reporter
}

// Start the YAML reporter and run until the context is cancelled.
func (r *YAMLReporter) Start(ctx context.Context) {
	go func() {
		for {
			select {
			case f := <-r.synchronizer:
				f()
			case <-ctx.Done():
				// This isn't safe and depends completely on that the coordinator
				// (arbiter) ensures no more calls will come to the reporter when
				// terminating this context. Should find a better solution
				close(r.synchronizer)
				for f := range r.synchronizer {
					f()
				}
				close(r.stopped)
				return
			}
		}
	}()
}

func (r *YAMLReporter) Op(mod, op string, res *op.Result, err error) {
	r.synchronizer <- func() {
		r.report.module(mod).addOp(op, res, err)
	}
}

func (r *YAMLReporter) LogErr(mod string, err error) {
	r.synchronizer <- func() {
		r.report.module(mod).addLogErr(err)
	}
}

func (r *YAMLReporter) LogTrigger(mod string, tr *report.TriggerReport[string]) {
	r.synchronizer <- func() {
		r.report.module(mod).addLogTrigger(tr)
	}
}

func (r *YAMLReporter) CPU(mod string, value float64) {
	r.synchronizer <- func() {
		r.report.module(mod).addCPU(value)
	}
}

func (r *YAMLReporter) CPUErr(mod string, err error) {
	r.synchronizer <- func() {
		r.report.module(mod).addCPUErr(err)
	}
}

func (r *YAMLReporter) CPUTrigger(mod string, tr *report.TriggerReport[float64]) {
	r.synchronizer <- func() {
		r.report.module(mod).addCPUTrigger(tr)
	}
}

func (r *YAMLReporter) RSS(mod string, value uint) {
	r.synchronizer <- func() {
		r.report.module(mod).addRSS(value)
	}
}

func (r *YAMLReporter) RSSErr(mod string, err error) {
	r.synchronizer <- func() {
		r.report.module(mod).addRSSErr(err)
	}
}

func (r *YAMLReporter) RSSTrigger(mod string, tr *report.TriggerReport[uint]) {
	r.synchronizer <- func() {
		r.report.module(mod).addRSSTrigger(tr)
	}
}

func (r *YAMLReporter) VMS(mod string, value uint) {
	r.synchronizer <- func() {
		r.report.module(mod).addVMS(value)
	}
}

func (r *YAMLReporter) VMSErr(mod string, err error) {
	r.synchronizer <- func() {
		r.report.module(mod).addVMSErr(err)
	}
}

func (r *YAMLReporter) VMSTrigger(mod string, tr *report.TriggerReport[uint]) {
	r.synchronizer <- func() {
		r.report.module(mod).addVMSTrigger(tr)
	}
}

func (r *YAMLReporter) MetricErr(mod, metric string, err error) {
	r.synchronizer <- func() {
		r.report.module(mod).addMetricErr(metric, err)
	}
}

func (r *YAMLReporter) MetricTrigger(mod, metric string, tr *report.TriggerReport[float64]) {
	r.synchronizer <- func() {
		r.report.module(mod).addMetricTrigger(metric, tr)
	}
}

func (r *YAMLReporter) Finalise() error {
	// Await synchronizer
	<-r.stopped

	r.report.End = time.Now()
	r.report.Duration = r.report.End.Sub(r.report.Start)

	file, err := os.Create(r.path)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := yaml.NewEncoder(file)
	encoder.SetIndent(2)
	return encoder.Encode(r.report)
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
