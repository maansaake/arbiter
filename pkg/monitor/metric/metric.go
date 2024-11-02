package metric

type metric struct {
}

func NewMetricMonitor() Metric {
	return &metric{}
}

func (m *metric) Pull() {}
