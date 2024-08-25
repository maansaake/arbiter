package log

import "github.com/go-logr/logr"

var logger logr.Logger

func Set(newLogger logr.Logger) {
	logger = newLogger
}

func Info(msg string, keysAndValues ...any) {
	logger.Info(msg, keysAndValues)
}

func V(verbosity int) logr.Logger {
	return logger.V(verbosity).WithCallDepth(-1)
}

func Error(err error, msg string, keysAndValues ...any) {
	logger.Error(err, msg, keysAndValues)
}

func WithName(name string) logr.Logger {
	return logger.WithName(name)
}

func WithCallDepth(depth int) logr.Logger {
	return logger.WithCallDepth(depth)
}

func WithValues(keysAndValues ...any) logr.Logger {
	return logger.WithValues(keysAndValues)
}
