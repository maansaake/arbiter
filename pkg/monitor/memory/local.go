package memory

import "github.com/shirou/gopsutil/v4/process"

type localMemory struct {
	process *process.Process
}

func NewLocalMemoryMonitor(pid int32) Memory {
	proc, err := process.NewProcess(pid)
	if err != nil {
		panic(err)
	}
	return &localMemory{process: proc}
}

func (m *localMemory) RSS() (uint, error) {
	memInfo, err := m.process.MemoryInfo()
	if err != nil {
		return 0, err
	}
	return uint(memInfo.RSS), nil
}

func (m *localMemory) VMS() (uint, error) {
	memInfo, err := m.process.MemoryInfo()
	if err != nil {
		return 0, err
	}
	return uint(memInfo.VMS), nil
}
