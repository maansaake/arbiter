package monitor

type Monitor struct {
	CPUMonitor
	MemoryMonitor
	MetricMonitor
	LogMonitor
}

type CPUMonitor interface{}
type MemoryMonitor interface{}
type MetricMonitor interface{}
type LogMonitor interface{}
