package interactive

import (
	"fmt"
	"io"
	"slices"
	"strings"

	"github.com/go-logr/logr"
)

// interactiveSink is a logr.LogSink that formats log messages as plain-text
// lines and writes them to the supplied io.Writer (typically a logCapture).
type interactiveSink struct {
	capture io.Writer
	name    string
	kvs     []any
	v       int
}

var _ logr.LogSink = &interactiveSink{}

// Init implements logr.LogSink.
func (s *interactiveSink) Init(_ logr.RuntimeInfo) {}

// Enabled implements logr.LogSink.
func (s *interactiveSink) Enabled(v int) bool { return v <= s.v }

// Info implements logr.LogSink.
func (s *interactiveSink) Info(_ int, msg string, keysAndValues ...any) {
	prefix := "[INFO]"
	if s.name != "" {
		prefix = fmt.Sprintf("[INFO] %s:", s.name)
	}

	fmt.Fprintf(s.capture, "%s %s%s\n", prefix, msg, formatKVs(append(s.kvs, keysAndValues...)))
}

// Error implements logr.LogSink.
func (s *interactiveSink) Error(err error, msg string, keysAndValues ...any) {
	prefix := "[ERROR]"
	if s.name != "" {
		prefix = fmt.Sprintf("[ERROR] %s:", s.name)
	}

	all := append([]any{"err", err.Error()}, append(s.kvs, keysAndValues...)...)
	fmt.Fprintf(s.capture, "%s %s%s\n", prefix, msg, formatKVs(all))
}

// WithValues implements logr.LogSink.
func (s *interactiveSink) WithValues(keysAndValues ...any) logr.LogSink {
	ns := *s
	ns.kvs = append(slices.Clone(s.kvs), keysAndValues...)
	return &ns
}

// WithName implements logr.LogSink.
func (s *interactiveSink) WithName(name string) logr.LogSink {
	ns := *s
	if ns.name == "" {
		ns.name = name
	} else {
		ns.name = ns.name + "/" + name
	}
	return &ns
}

// WithCallDepth implements logr.LogSink.
func (s *interactiveSink) WithCallDepth(_ int) logr.LogSink { return s }

// formatKVs formats key-value pairs as " key=value key=value…".
func formatKVs(kvs []any) string {
	if len(kvs) == 0 {
		return ""
	}

	var b strings.Builder
	for i := 0; i+1 < len(kvs); i += 2 {
		fmt.Fprintf(&b, " %v=%v", kvs[i], kvs[i+1])
	}

	return b.String()
}
