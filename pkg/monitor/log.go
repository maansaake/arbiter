package monitor

type log struct{}

func NewLogMonitor() Log {
	return &log{}
}
