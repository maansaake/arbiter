package zerologr

import (
	"fmt"
	"os"

	"github.com/go-logr/logr"
	"github.com/rs/zerolog"
)

var logger logr.Logger //nolint:gochecknoglobals // package-level logger for this facade

// SetLogger automatically adds calldepth to account for this facade.
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

type sink struct {
	logger    *zerolog.Logger
	v         int
	name      string
	callDepth int
}

// Opts provides options to tweak the underlying zerolog logsink.
type Opts struct {
	Console bool
	Caller  bool
	V       int
}

var (
	// VFieldName can be set to change the name of the verbosity field. An empty
	// value means the field is not emitted in log events.
	VFieldName string //nolint:gochecknoglobals // exported config var, set via SetVFieldName
	// NameFieldName can be set to change the name of the logger name field. This
	// field is only included if the logger has been given a name via WithName().
	NameFieldName = "logger" //nolint:gochecknoglobals // exported config var
)

// SetVFieldName sets the verbosity field name used in log output.
func SetVFieldName(name string) {
	VFieldName = name
}

// Some global zerolog overrides.
func init() { //nolint:gochecknoinits // initializes zerolog global settings at package load
	// Only info level as debug and trace levels are omitted. There is only Info/
	// Error logs in the logr world.
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	// Short and sweet.
	zerolog.TimestampFieldName = "t" //nolint:reassign // configuring zerolog global timestamp field name
}

// New creates a new logr.Logger with the input options. Console logging uses
// the pretty zerolog output, not meant for production.
func New(opts *Opts) logr.Logger {
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	if opts.Caller {
		logger = logger.With().Caller().Logger()
	}

	if opts.Console {
		logger = logger.Output(zerolog.ConsoleWriter{Out: os.Stdout, NoColor: true, TimeFormat: "2006-01-02 15:04:05"})
	}

	return logr.New(&sink{
		logger: &logger,
		v:      opts.V,
	})
}

const callDepthAdjustment = 2

// Init implements logr.LogSink.
func (s *sink) Init(info logr.RuntimeInfo) {
	s.callDepth = info.CallDepth + callDepthAdjustment
}

// Enabled implements logr.LogSink.
func (s *sink) Enabled(level int) bool {
	return level <= s.v
}

// Error implements logr.LogSink.
func (s *sink) Error(err error, msg string, keysAndValues ...any) {
	e := s.logger.Err(err)
	s.msg(e, msg, keysAndValues...)
}

// Info implements logr.LogSink.
func (s *sink) Info(level int, msg string, keysAndValues ...any) {
	e := s.logger.Info()

	if VFieldName != "" {
		e.Int(VFieldName, level)
	}

	s.msg(e, msg, keysAndValues...)
}

func (s *sink) msg(e *zerolog.Event, msg string, keysAndValues ...any) {
	if s.name != "" {
		e.Str(NameFieldName, s.name)
	}

	e.CallerSkipFrame(s.callDepth)
	e.Fields(keysAndValues)
	e.Msg(msg)
}

// WithName implements logr.LogSink.
func (s *sink) WithName(name string) logr.LogSink {
	ns := *s
	if s.name != "" {
		ns.name = fmt.Sprintf("%s/%s", s.name, name)
	} else {
		ns.name = name
	}
	return &ns
}

// WithValues implements logr.LogSink.
func (s *sink) WithValues(keysAndValues ...any) logr.LogSink {
	ns := *s
	nl := s.logger.With().Fields(keysAndValues).Logger()
	ns.logger = &nl
	return &ns
}

// WithValues implements logr.CallDepthLogSink.
func (s *sink) WithCallDepth(level int) logr.LogSink {
	ns := *s
	ns.callDepth += level
	return &ns
}
