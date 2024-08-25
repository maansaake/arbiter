package log

import (
	"fmt"
	"os"

	"github.com/go-logr/logr"
	"github.com/rs/zerolog"
)

type CustomSink struct {
	zerologger *zerolog.Logger

	verbosity          int
	verbosityFieldName string

	name          string
	nameFieldName string

	console bool
	caller  bool
	depth   int
}

type Options struct {
	Verbosity          int
	VerbosityFieldName string
	NameFieldName      string
	Console            bool
	Caller             bool
}

func NewSink(opts *Options) logr.LogSink {
	zerologger := zerolog.New(os.Stdout).
		With().
		Timestamp().
		Logger().
		Level(zerolog.InfoLevel)

	if opts.Caller {
		zerologger = zerologger.With().Caller().Logger()
	}

	if opts.Console {
		zerologger = zerologger.Output(zerolog.ConsoleWriter{
			Out:     os.Stdout,
			NoColor: true,
		})
	}

	computedNameFieldName := "logger"
	if opts.NameFieldName != "" {
		computedNameFieldName = opts.NameFieldName
	}
	return &CustomSink{
		zerologger:         &zerologger,
		verbosity:          opts.Verbosity,
		verbosityFieldName: opts.VerbosityFieldName,
		nameFieldName:      computedNameFieldName,
		console:            opts.Console,
		caller:             opts.Caller,
	}
}

// Init implements logr.LogSink.
func (c *CustomSink) Init(info logr.RuntimeInfo) {
	c.depth = info.CallDepth + 3
}

// Enabled implements logr.LogSink.
func (c *CustomSink) Enabled(level int) bool {
	return level <= c.verbosity
}

// Error implements logr.LogSink.
func (c *CustomSink) Error(err error, msg string, keysAndValues ...any) {
	c.msg(c.zerologger.Error().Err(err), msg, keysAndValues...)
}

// Info implements logr.LogSink.
func (c *CustomSink) Info(level int, msg string, keysAndValues ...any) {
	event := c.zerologger.Info()
	if c.verbosityFieldName != "" {
		event.Int(c.verbosityFieldName, level)
	}
	c.msg(event, msg, keysAndValues...)
}

// WithName implements logr.LogSink.
func (c *CustomSink) WithName(name string) logr.LogSink {
	cn := *c

	if cn.name == "" {
		cn.name = name
	} else {
		cn.name = fmt.Sprintf("%s/%s", cn.name, name)
	}

	zl := cn.zerologger.With().Str(cn.nameFieldName, cn.name).Logger()
	cn.zerologger = &zl

	return &cn
}

// WithValues implements logr.LogSink.
func (c *CustomSink) WithValues(keysAndValues ...any) logr.LogSink {
	cn := *c
	zl := cn.zerologger.With().Fields(keysAndValues).Logger()
	cn.zerologger = &zl
	return &cn
}

func (c *CustomSink) WithCallDepth(depth int) logr.LogSink {
	cn := *c
	cn.depth = cn.depth + depth
	return &cn
}

func (c *CustomSink) msg(event *zerolog.Event, msg string, keysAndValues ...any) {
	event.CallerSkipFrame(c.depth)
	event.Fields(keysAndValues)
	event.Msg(msg)
}
