package monitor

type logFile struct{}

func NewLogFileMonitor() Log {
	return &logFile{}
}
