package traffic

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/go-logr/logr"
	"github.com/maansaake/arbiter/pkg/module"
	"github.com/maansaake/arbiter/pkg/report"
)

var (
	ErrNoOpsToSchedule = errors.New("there were no operations to schedule")
	ErrZeroRate        = errors.New("operation has a zero rate")
	ErrCleanupTimeout  = errors.New("cleanup timed out")
	ErrRateIssue       = errors.New("rate issue")
)

const (
	DefaultWorkerLimit = 10

	defaultSampleIntervalSeconds = 10
	cleanupTimeout               = 5 * time.Second
	minRateForDefaultSample      = 30
	defaultSampleTolerancePerc   = 0.05
)

// Opts configures a Scheduler.
type Opts struct {
	// Logger is used for traffic scheduler logs. Defaults to a discard logger if not set.
	Logger logr.Logger
	// WorkerLimit is the maximum number of concurrent workers per workload. Defaults to DefaultWorkerLimit.
	WorkerLimit int
	// SampleTolerancePerc is the tolerance percentage used when comparing sampled rates in tests.
	// Defaults to 0.05 (5%).
	SampleTolerancePerc float64
}

// Scheduler runs traffic against registered modules.
type Scheduler interface {
	// Run starts traffic generation for the given modules, reporting results to reporter.
	// It is asynchronous: it returns once the goroutines are launched and monitors ctx
	// to stop gracefully when it is cancelled.
	Run(ctx context.Context, metadata module.Metadata, reporter report.Reporter) error
	// Stop waits for all workloads to finish after the context passed to Run is cancelled.
	Stop() error
}

type scheduler struct {
	logger              logr.Logger
	workerLimit         int
	sampleTolerancePerc float64

	workloads []*workload
	stopChan  chan *workload
}

// New creates a Scheduler with the given options. A nil opts uses all defaults.
func New(opts *Opts) Scheduler {
	if opts == nil {
		opts = &Opts{}
	}
	if opts.WorkerLimit == 0 {
		opts.WorkerLimit = DefaultWorkerLimit
	}
	if opts.SampleTolerancePerc == 0 {
		opts.SampleTolerancePerc = defaultSampleTolerancePerc
	}
	return &scheduler{
		logger:              opts.Logger,
		workerLimit:         opts.WorkerLimit,
		sampleTolerancePerc: opts.SampleTolerancePerc,
	}
}

// Run traffic for the input modules using their exposed operations. Traffic
// generation will make operation calls at the specified rates and report
// problems to the reporter. Run() is asynchronous and returns once the main
// go-routine has been started. Run() will monitor the context's done channel
// and stop gracefully once it's closed.
func (s *scheduler) Run(
	ctx context.Context,
	metadata module.Metadata,
	reporter report.Reporter,
) error {
	s.logger.Info("Running traffic generator")

	s.workloads = make([]*workload, 0, len(metadata))
	for _, meta := range metadata {
		for _, op := range meta.Ops() {
			if op.Disabled {
				s.logger.Info("Skipping disabled operation", "mod", meta.Name(), "op", op.Name)
				continue
			}

			if op.Rate == 0 {
				return fmt.Errorf("%w: %s", ErrZeroRate, op.Name)
			}

			s.workloads = append(s.workloads, &workload{
				workerLimit: s.workerLimit,
				statLock:    &sync.Mutex{},
				mod:         meta.Name(),
				op:          op,
				reporter:    reporter,
				logger:      s.logger,
			})
		}
	}

	if len(s.workloads) == 0 {
		return ErrNoOpsToSchedule
	}

	// Create stop channel that workloads will report to when stopping.
	s.stopChan = make(chan *workload, len(s.workloads))
	for _, wl := range s.workloads {
		wl.stopChan = s.stopChan
	}

	// Run the workloads in separate go-routines, each runs until context is done.
	for _, wl := range s.workloads {
		go wl.run(ctx)
	}

	return nil
}

// Stop waits for all workloads to finish and returns any error encountered.
func (s *scheduler) Stop() error {
	s.logger.Info("Stopping traffic generator", "workload_count", len(s.workloads))

	stopCount := 0
	for {
		select {
		case <-time.After(cleanupTimeout):
			s.logger.Error(ErrCleanupTimeout, "Cleanup timed out after "+cleanupTimeout.String())
			return ErrCleanupTimeout
		case wl := <-s.stopChan:
			s.logger.Info("Workload stopped", "mod", wl.mod, "op", wl.op.Name)
			stopCount++
			if stopCount == len(s.workloads) {
				s.logger.Info("All workloads have stopped")
				return nil
			}
		}
	}
}

func getSampleInterval(op *module.Op) time.Duration {
	if op.Rate < minRateForDefaultSample {
		// Minimum 5 samples, this should be a super corner case. Add some time
		// to allow the 5th invocation to fire.
		return time.Minute/time.Duration(op.Rate)*5 + 250*time.Millisecond
	}

	return defaultSampleIntervalSeconds * time.Second
}
