package memory

type localMemory struct {
}

func NewLocalMemoryMonitor() Memory {
	return &localMemory{}
}

func (m *localMemory) Read() uint {
	return 0
}
