package yaml

import (
	"time"

	moduleop "tres-bon.se/arbiter/pkg/module/op"
	"tres-bon.se/arbiter/pkg/report"
)

type (
	module struct {
		CPU        *cpu                  `yaml:"cpu,omitempty"`
		Mem        *mem                  `yaml:"mem,omitempty"`
		Log        *log                  `yaml:"log,omitempty"`
		Metric     *metric               `yaml:"metric,omitempty"`
		Operations map[string]*operation `yaml:"operation"`
	}
	cpu struct {
		readings uint                             `yaml:"-"`
		total    float64                          `yaml:"-"`
		Average  float64                          `yaml:"average"`
		High     float64                          `yaml:"high"`
		Low      float64                          `yaml:"low"`
		Errs     []string                         `yaml:"errors"`
		Triggers []*report.TriggerReport[float64] `yaml:"triggers"`
	}
	mem struct {
		RSS *rss `yaml:"rss"`
		VMS *vms `yaml:"vms"`
	}
	log struct {
		Errs     []string                        `yaml:"errors"`
		Triggers []*report.TriggerReport[string] `yaml:"triggers"`
	}
	metric struct {
		Errs     map[string][]string                         `yaml:"errors"`
		Triggers map[string][]*report.TriggerReport[float64] `yaml:"triggers"`
	}
	rss struct {
		readings uint                          `yaml:"-"`
		total    uint                          `yaml:"-"`
		Average  uint                          `yaml:"average"`
		High     uint                          `yaml:"high"`
		Low      uint                          `yaml:"low"`
		Errs     []string                      `yaml:"errors"`
		Triggers []*report.TriggerReport[uint] `yaml:"triggers"`
	}
	vms struct {
		readings uint                          `yaml:"-"`
		total    uint                          `yaml:"-"`
		Average  uint                          `yaml:"average"`
		High     uint                          `yaml:"high"`
		Low      uint                          `yaml:"low"`
		Errs     []string                      `yaml:"errors"`
		Triggers []*report.TriggerReport[uint] `yaml:"triggers"`
	}
	operation struct {
		Executions uint             `yaml:"executions"`
		OK         uint             `yaml:"ok"`
		NOK        uint             `yaml:"nok"`
		Timing     *operationTiming `yaml:"timing"`
		Errors     []string         `yaml:"errors,omitempty"`
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
		Mem: &mem{
			RSS: &rss{},
			VMS: &vms{},
		},
		Log: &log{},
		Metric: &metric{
			Errs:     make(map[string][]string),
			Triggers: make(map[string][]*report.TriggerReport[float64]),
		},
		Operations: make(map[string]*operation),
	}
}

func (m *module) addOp(name string, res *moduleop.Result, err error) {
	op, ok := m.Operations[name]
	if !ok {
		op = &operation{
			Timing: &operationTiming{},
		}
		m.Operations[name] = op
	}

	op.Executions++

	if err != nil {
		op.NOK++
		op.Errors = append(op.Errors, err.Error())
	} else {
		op.OK++
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
	m.Log.Errs = append(m.Log.Errs, err.Error())
}

func (m *module) addLogTrigger(tr *report.TriggerReport[string]) {
	m.Log.Triggers = append(m.Log.Triggers, tr)
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
	m.CPU.Errs = append(m.CPU.Errs, err.Error())
}

func (m *module) addCPUTrigger(tr *report.TriggerReport[float64]) {
	m.CPU.Triggers = append(m.CPU.Triggers, tr)
}

func (m *module) addRSS(value uint) {
	if m.Mem.RSS.readings == 0 {
		m.Mem.RSS.High = value
		m.Mem.RSS.Low = value
	}

	m.Mem.RSS.readings++

	if value > m.Mem.RSS.High {
		m.Mem.RSS.High = value
	}

	if value < m.Mem.RSS.Low {
		m.Mem.RSS.Low = value
	}

	m.Mem.RSS.total += value
	m.Mem.RSS.Average = m.Mem.RSS.total / m.Mem.RSS.readings
}

func (m *module) addRSSErr(err error) {
	m.Mem.RSS.Errs = append(m.Mem.RSS.Errs, err.Error())
}

func (m *module) addRSSTrigger(tr *report.TriggerReport[uint]) {
	m.Mem.RSS.Triggers = append(m.Mem.RSS.Triggers, tr)
}

func (m *module) addVMS(value uint) {
	if m.Mem.VMS.readings == 0 {
		m.Mem.VMS.High = value
		m.Mem.VMS.Low = value
	}

	m.Mem.VMS.readings++

	if value > m.Mem.VMS.High {
		m.Mem.VMS.High = value
	}

	if value < m.Mem.VMS.Low {
		m.Mem.VMS.Low = value
	}

	m.Mem.VMS.total += value
	m.Mem.VMS.Average = m.Mem.VMS.total / m.Mem.VMS.readings
}

func (m *module) addVMSErr(err error) {
	m.Mem.VMS.Errs = append(m.Mem.VMS.Errs, err.Error())
}

func (m *module) addVMSTrigger(tr *report.TriggerReport[uint]) {
	m.Mem.VMS.Triggers = append(m.Mem.VMS.Triggers, tr)
}

func (m *module) addMetricErr(metric string, err error) {
	m.Metric.Errs[metric] = append(m.Metric.Errs[metric], err.Error())
}

func (m *module) addMetricTrigger(metric string, tr *report.TriggerReport[float64]) {
	m.Metric.Triggers[metric] = append(m.Metric.Triggers[metric], tr)
}
