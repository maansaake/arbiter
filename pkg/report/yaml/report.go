package yamlreport

import (
	"time"

	"github.com/maansaake/arbiter/pkg/module"
)

type (
	// Report is the YAML report struct. It contains all the information about the execution of the modules and their operations.
	Report struct {
		Start    time.Time                `yaml:"start"`
		End      time.Time                `yaml:"end"`
		Duration time.Duration            `yaml:"duration"`
		Modules  map[string]*ModuleReport `yaml:"modules"`
	}
	// ModuleReport contains the report information for a module. It contains the operations and their respective reports.
	ModuleReport struct {
		Operations map[string]*OperationDetails `yaml:"operation"`
	}
	// OperationDetails contains the report information for an operation.
	OperationDetails struct {
		Executions uint             `yaml:"executions"`
		OK         uint             `yaml:"ok"`
		NOK        uint             `yaml:"nok"`
		Timing     *OperationTiming `yaml:"timing"`
	}
	// OperationTiming contains the timing information for an operation.
	OperationTiming struct {
		Longest  time.Duration `yaml:"longest"`
		Shortest time.Duration `yaml:"shortest"`
		Average  time.Duration `yaml:"average"`
		total    time.Duration `yaml:"-"`
		// Needed since executions count failures that do not count towards timing
		// stats.
		count int64 `yaml:"-"`
	}
)

func newModuleReport() *ModuleReport {
	return &ModuleReport{
		Operations: make(map[string]*OperationDetails),
	}
}

func (m *ModuleReport) addOp(name string, res *module.Result, err error) {
	op, ok := m.Operations[name]
	if !ok {
		op = &OperationDetails{
			Timing: &OperationTiming{},
		}
		m.Operations[name] = op
	}

	op.Executions++

	if err != nil {
		op.NOK++
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
