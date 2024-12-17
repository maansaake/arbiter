package log

import (
	"context"
	"errors"
)

type (
	// The Log interface exposes just one method, returning a stream of the log
	// being monitored. The log could be a file, a stream or other.
	Log interface {
		Stream(context.Context, LogHandler) error
	}
	// A LogHandler accepts log input and handles it in some way.
	LogHandler func(string, error)
)

var ErrStopped = errors.New("log handler is stopped")
