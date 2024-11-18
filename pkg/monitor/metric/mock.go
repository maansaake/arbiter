package metric

type MetricMock struct {
	Metric

	onRead func()
	scrape []byte
	err    error
}

func NewMetricMock(scrape []byte, err error) *MetricMock {
	return &MetricMock{
		scrape: scrape,
		err:    err,
	}
}

func (m *MetricMock) Pull() ([]byte, error) {
	if m.onRead != nil {
		m.onRead()
	}
	return m.scrape, m.err
}

func (m *MetricMock) LatestRawMetrics() []byte {
	if m.onRead != nil {
		m.onRead()
	}
	return m.scrape
}

func (m *MetricMock) SetScrape(newScrape []byte) {
	m.scrape = newScrape
}

func (m *MetricMock) SetErr(newErr error) {
	m.err = newErr
}

func (m *MetricMock) SetOnRead(onRead func()) {
	m.onRead = onRead
}
