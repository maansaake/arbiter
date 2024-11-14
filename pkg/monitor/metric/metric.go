package metric

type metric struct {
	endpoint string
}

func NewMetricMonitor(endpoint string) Metric {
	return &metric{
		endpoint: endpoint,
	}
}

func (m *metric) Pull() {}
