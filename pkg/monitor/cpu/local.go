package cpu

import (
	"github.com/shirou/gopsutil/v4/process"
)

type localCpu struct {
	proc *process.Process
}

func NewLocalCPUMonitor(pid int32) CPU {
	proc, err := getProc(pid)
	if err != nil {
		panic(err)
	}

	// Initial ignored call to avoid 0 return for first reading.
	_, _ = proc.Percent(0)

	return &localCpu{proc: proc}
}

func (c *localCpu) Read() (float64, error) {
	// Percent(0) returns the CPU percentage since the last reading.
	return c.proc.Percent(0)
}

func getProc(pid int32) (*process.Process, error) {
	return process.NewProcess(pid)
}
