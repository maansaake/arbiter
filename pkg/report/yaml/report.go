package yaml

import (
	"time"

	"tres-bon.se/arbiter/pkg/module"
)

type (
	// Struct for the YAML yamlReport. The entire thing is marshaled into the final
	// file.
	yamlReport struct {
		Start    time.Time                `yaml:"start"`
		End      time.Time                `yaml:"end"`
		Duration time.Duration            `yaml:"duration"`
		Modules  map[string]*moduleReport `yaml:"modules"`
	}

	moduleReport struct {
		Operations map[string]*operation `yaml:"operation"`
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

func newModule() *moduleReport {
	return &moduleReport{
		Operations: make(map[string]*operation),
	}
}

func (m *moduleReport) addOp(name string, res *module.Result, err error) {
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

		op.Timing.Average = op.Timing.total / time.Duration(op.Timing.count)
	}
}
