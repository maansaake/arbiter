package yaml

import (
	"context"
	"os"
	"time"

	"github.com/maansaake/arbiter/pkg/module"
	"github.com/maansaake/arbiter/pkg/report"
	"github.com/trebent/zerologr"
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
	}
	// Reporter implements the reporter interface. //nolint:revive // exported type name stutter is intentional for clarity.
	Reporter struct {
		// The final path of the YAML report.
		path string
		// The YAML report.
		report *Report
		// Synchronizer channel to limit access to the report to 1 thread. Also
		// speeds up calls to the reporter interface.
		synchronizer chan func()
		stopped      chan struct{}
	}
)

var _ report.Reporter = &Reporter{}

const yamlIndent = 2

// New creates a new YAML reporter.
func New(opts *Opts) *Reporter {
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

	reporter := &Reporter{
		report: &Report{
			Start:   start,
			Modules: make(map[string]*ModuleReport),
		},
		path:         opts.Path,
		synchronizer: make(chan func(), buffer),
		stopped:      make(chan struct{}),
	}

	return reporter
}

// Start the YAML reporter and run until the context is cancelled.
func (r *Reporter) Start(ctx context.Context) {
	zerologr.Info("Starting reporter")

	go func() {
		for {
			select {
			case f := <-r.synchronizer:
				f()
			case <-ctx.Done():
				zerologr.Info("Reporter context closed, flushing synchronizer", "len", len(r.synchronizer))

				// This isn't safe and depends completely on that the coordinator
				// (arbiter) ensures no more calls will come to the reporter when
				// terminating this context. Should find a better solution
				close(r.synchronizer)
				for f := range r.synchronizer {
					f()
				}
				zerologr.Info("Synchronizer flushed, stopping reporter")
				close(r.stopped)
				return
			}
		}
	}()
}

func (r *Reporter) Op(mod, op string, res *module.Result, err error) {
	r.synchronizer <- func() {
		r.report.module(mod).addOp(op, res, err)
	}
}

func (r *Reporter) Finalise() error {
	// Await synchronizer, no value expected
	<-r.stopped
	zerologr.Info("Synchronizer stopped, writing report")

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
