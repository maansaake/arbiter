package yamlreport

import (
	"context"
	"errors"
	"os"
	"time"

	"github.com/go-logr/logr"
	"github.com/maansaake/arbiter/pkg/module"
	"github.com/maansaake/arbiter/pkg/report"
	"gopkg.in/yaml.v3"
)

type (
	// Opts contains options for the YAML reporter.
	Opts struct {
		// Start time to set in the report. If left empty a start time is set
		// when calling `New()`.
		Start time.Time
		// The final path of the YAML report.
		Path string
		// The buffer size sets the number of buffered report calls that are yet
		// to be handled. Values < 1 will be ignored.
		Buffer int
		// Logger is the logger used for info-level logging by the reporter.
		Logger logr.Logger
		// ErrorLogger is a logger for the reporter to log errors to.
		ErrorLogger logr.Logger
	}
	// reporter implements the reporter interface. //nolint:revive // exported type name stutter is intentional for clarity.
	reporter struct {
		// The final path of the YAML report.
		path string
		// The YAML report.
		report *Report
		// logger is used for info-level logging.
		logger logr.Logger
		// errorLogger is used to log errors from failed operations.
		errorLogger logr.Logger
		// Synchronizer channel to limit access to the report to 1 thread. Also
		// speeds up calls to the reporter interface.
		synchronizer chan func()
		stopped      chan struct{}
	}
)

var _ report.Reporter = &reporter{}

const yamlIndent = 2

// New creates a new YAML reporter.
func New(opts *Opts) (report.Reporter, error) {
	if opts.Logger.GetSink() == nil {
		return nil, errors.New("logger must be provided")
	}

	var start time.Time
	var buffer int
	if opts.Buffer > 0 {
		buffer = opts.Buffer
	} else {
		buffer = 100
	}

	if opts.Start.IsZero() {
		start = time.Now()
	} else {
		start = opts.Start
	}

	reporter := &reporter{
		report: &Report{
			Start:   start,
			Modules: make(map[string]*ModuleReport),
		},
		logger:       opts.Logger,
		errorLogger:  opts.ErrorLogger,
		path:         opts.Path,
		synchronizer: make(chan func(), buffer),
		stopped:      make(chan struct{}),
	}

	return reporter, nil
}

// Start the YAML reporter and run until the context is cancelled.
func (r *reporter) Start(ctx context.Context) {
	r.logger.Info("Starting reporter")

	go func() {
		for {
			select {
			case f := <-r.synchronizer:
				f()
			case <-ctx.Done():
				r.logger.Info("Reporter context closed, flushing synchronizer", "len", len(r.synchronizer))

				// TODO: test buffer emptying
			out:
				// Empty the synchronizer buffer, up to 100 items, if not empty.
				for range 100 {
					select {
					case f := <-r.synchronizer:
						f()
					default:
						break out
					}
				}
					r.logger.Info("Synchronizer flushed, stopping reporter")
				close(r.stopped)
				return
			}
		}
	}()
}

func (r *reporter) ReportError(_ error) {
	// no-op
}

func (r *reporter) ReportOp(mod, op string, res *module.Result, err error) {
	r.synchronizer <- func() {
		r.report.module(mod).addOp(op, res, err)
		if err != nil {
			r.errorLogger.Error(err, "Error in operation", "mod", mod, "op", op)
		}
	}
}

func (r *reporter) Finalise() error {
	// Await synchronizer, no value expected
	<-r.stopped
	r.logger.Info("Synchronizer stopped, writing report")

	r.report.End = time.Now()
	r.report.Duration = r.report.End.Sub(r.report.Start)

	file, err := os.Create(r.path)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := yaml.NewEncoder(file)
	defer encoder.Close()
	encoder.SetIndent(yamlIndent)
	return encoder.Encode(r.report)
}

/*INTERNAL*/

func (r *Report) module(mod string) *ModuleReport {
	m, ok := r.Modules[mod]
	if !ok {
		m = newModuleReport()
		r.Modules[mod] = m
	}

	return m
}
