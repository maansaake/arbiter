package log

type logFile struct{}

func NewLogFileMonitor() Log {
	return &logFile{}
}
