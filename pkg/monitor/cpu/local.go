package cpu

import (
	"github.com/shirou/gopsutil/v4/process"
)

type localCpu struct {
	process *process.Process
}

func NewLocalCPUMonitor(pid int32) CPU {
	proc, err := process.NewProcess(pid)
	if err != nil {
		panic(err)
	}

	// Initial ignored call to avoid 0 return for first reading.
	_, _ = proc.Percent(0)

	return &localCpu{process: proc}
}

func (c *localCpu) Read() (float64, error) {
	return c.process.Percent(0)
}
