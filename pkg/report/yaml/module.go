package yaml

import (
	"time"

	moduleop "tres-bon.se/arbiter/pkg/module/op"
	"tres-bon.se/arbiter/pkg/report"
)

type (
	module struct {
		CPU        *cpu                  `yaml:"cpu"`
		Mem        *mem                  `yaml:"mem"`
		Log        *log                  `yaml:"log"`
		Metric     *metric               `yaml:"metric"`
		Operations map[string]*operation `yaml:"operation"`
	}
	cpu struct {
		errs     []error                          `yaml:"errors"`
		triggers []*report.TriggerReport[float64] `yaml:"triggers"`
		readings uint                             `yaml:"-"`
		total    float64                          `yaml:"-"`
		Average  float64                          `yaml:"average"`
		High     float64                          `yaml:"high"`
		Low      float64                          `yaml:"low"`
	}
	mem struct {
		RSS *rss `yaml:"rss"`
		VMS *vms `yaml:"vms"`
	}
	log struct {
		errs     []error                         `yaml:"errors"`
		triggers []*report.TriggerReport[string] `yaml:"triggers"`
	}
	metric struct {
		errs     []error                          `yaml:"errors"`
		triggers []*report.TriggerReport[float64] `yaml:"triggers"`
	}
	rss struct {
		errs     []error                          `yaml:"errors"`
		triggers []*report.TriggerReport[float64] `yaml:"triggers"`
		readings uint                             `yaml:"-"`
		total    float64                          `yaml:"-"`
		Average  float64                          `yaml:"average"`
		High     float64                          `yaml:"high"`
		Low      float64                          `yaml:"low"`
	}
	vms struct {
		errs     []error                          `yaml:"errors"`
		triggers []*report.TriggerReport[float64] `yaml:"triggers"`
		readings uint                             `yaml:"-"`
		total    float64                          `yaml:"-"`
		Average  float64                          `yaml:"average"`
		High     float64                          `yaml:"high"`
		Low      float64                          `yaml:"low"`
	}
	operation struct {
		Executions uint             `yaml:"executions"`
		Timing     *operationTiming `yaml:"timing"`
		Errors     []string         `yaml:"errors"`
	}
	operationTiming struct {
		Longest  time.Duration `yaml:"longest"`
		Shortest time.Duration `yaml:"shortest"`
		Average  time.Duration `yaml:"average"`
		total    time.Duration `yaml:"-"`
		// Needed since executions count failures that do not count towards timing
		// stats.
		count int64 `yaml:"-"`
	}
)

func newModule() *module {
	return &module{
		CPU: &cpu{},
		Mem: &mem{},
		Log: &log{
			errs:     make([]error, 0),
			triggers: make([]*report.TriggerReport[string], 0),
		},
		Operations: make(map[string]*operation),
	}
}

func (m *module) addOp(res *moduleop.Result, err error) {
	op, ok := m.Operations[res.Name]
	if !ok {
		op = &operation{
			Timing: &operationTiming{},
			Errors: make([]string, 0),
		}
		m.Operations[res.Name] = op
	}

	op.Executions++

	if err != nil {
		op.Errors = append(op.Errors, err.Error())
	} else {
		if op.Timing.count == 0 {
			op.Timing.Longest = res.Duration
			op.Timing.Shortest = res.Duration
		}
		op.Timing.count++

		if res.Duration > op.Timing.Longest {
			op.Timing.Longest = res.Duration
		}
		if res.Duration < op.Timing.Shortest {
			op.Timing.Shortest = res.Duration
		}

		op.Timing.total += res.Duration
		//nolint:gosec
		op.Timing.Average = op.Timing.total / time.Duration(op.Timing.count)
	}
}

func (m *module) addLogErr(err error) {
	m.Log.errs = append(m.Log.errs, err)
}

func (m *module) addLogTrigger(tr *report.TriggerReport[string]) {
	m.Log.triggers = append(m.Log.triggers, tr)
}

func (m *module) addCPU(val float64) {
	if m.CPU.readings == 0 {
		m.CPU.High = val
		m.CPU.Low = val
	}
	m.CPU.readings++

	if val > m.CPU.High {
		m.CPU.High = val
	}

	if val < m.CPU.Low {
		m.CPU.Low = val
	}

	m.CPU.total += val
	m.CPU.Average = m.CPU.total / float64(m.CPU.readings)
}

func (m *module) addCPUErr(err error) {
}

func (m *module) addCPUTrigger(tr *report.TriggerReport[float64]) {
}

func (m *module) addRSS(value uint) {
}

func (m *module) addRSSErr(err error) {
}

func (m *module) addRSSTrigger(tr *report.TriggerReport[uint]) {
}

func (m *module) addVMS(value uint) {
}

func (m *module) addVMSErr(err error) {
}

func (m *module) addVMSTrigger(tr *report.TriggerReport[uint]) {
}

func (m *module) addMetricErr(err error) {
}

func (m *module) addMetricTrigger(tr *report.TriggerReport[float64]) {
}
