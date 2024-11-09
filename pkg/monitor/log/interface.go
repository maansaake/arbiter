package log

// The Log interface exposes just one method, returning a stream of the log
// being monitored. The log could be a file, a stream or other.
type Log interface {
	Stream(LogHandler)
}

// A LogHandler accepts log input and handles it in some way.
type LogHandler func(string)
