package log

type logFile struct {
	file    string
	handler LogHandler
}

func NewLogFileMonitor(file string) Log {
	return &logFile{file: file}
}

func (l *logFile) Stream(handler LogHandler) {
	l.handler = handler
}
