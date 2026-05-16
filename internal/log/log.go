package abtrlog

import (
	"os"

	"github.com/go-logr/logr"
	"github.com/trebent/zerologr"
)

type Opts struct {
	Verbosity    int
	ErrorLogPath string
	InfoLogPath  string
}

const (
	DefaultInfoLogPath  = "info.log"
	DefaultErrorLogPath = "error.log"
)

//nolint:gochecknoglobals // package-level state for loggers
var logger logr.Logger

// Setup sets up the loggers based on the provided options. It sets the package
// logger and returns an error logger for use by an arbiter reporter.
func Setup(opts *Opts) (logr.Logger, error) {
	if opts == nil {
		opts = &Opts{
			Verbosity:    0,
			ErrorLogPath: "error.log",
			InfoLogPath:  "info.log",
		}
	}

	if opts.Verbosity < 0 {
		opts.Verbosity = 0
	}

	if opts.ErrorLogPath == "" {
		opts.ErrorLogPath = "error.log"
	}

	// Set up the logger based on the provided options.
	errorFile, err := os.Create(opts.ErrorLogPath)
	if err != nil {
		return logr.Logger{}, err
	}

	if opts.InfoLogPath == "" {
		logger = logr.Discard()
	} else {
		//nolint:govet // shad
		infoFile, err := os.Create(opts.InfoLogPath)
		if err != nil {
			return logr.Logger{}, err
		}

		logger = zerologr.New(&zerologr.Opts{
			Output:  infoFile,
			Console: true,
			Caller:  true,
			V:       opts.Verbosity,
		})
	}

	errorLogger := zerologr.New(&zerologr.Opts{
		Output:  errorFile,
		Console: true,
		Caller:  false,
		V:       opts.Verbosity,
	})

	return errorLogger, nil
}

// GetLogger returns the package logger for use by other components. Setup should be called first
// to initialize the logger before calling this function.
func GetLogger() logr.Logger {
	return logger
}
