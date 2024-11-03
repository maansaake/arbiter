package cpu

import "github.com/shirou/gopsutil/v4/process"

type localCpu struct {
	process *process.Process
}

func NewLocalCPUMonitor(pid int32) CPU {
	proc, err := process.NewProcess(pid)
	if err != nil {
		panic(err)
	}

	return &localCpu{process: proc}
}

func (c *localCpu) Read() (float64, error) {
	return c.process.CPUPercent()
}
