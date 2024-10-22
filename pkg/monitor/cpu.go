package monitor

type localCpu struct{}

func NewLocalCPUMonitor() CPU {
	return &localCpu{}
}

func (c *localCpu) Read() float32 {
	return 0.0
}
