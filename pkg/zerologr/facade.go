package zerologr

import (
	"github.com/go-logr/logr"
)

var (
	logger logr.Logger
)

// Automatically adds calldepth to account for this facade.
func SetLogger(newLogger logr.Logger) {
	logger = newLogger.WithCallDepth(1)
}

func V(level int) logr.Logger {
	return logger.V(level).WithCallDepth(-1)
}

func Info(msg string, keysAndValues ...any) {
	logger.Info(msg, keysAndValues...)
}

func Error(err error, msg string, keysAndValues ...any) {
	logger.Error(err, msg, keysAndValues...)
}
