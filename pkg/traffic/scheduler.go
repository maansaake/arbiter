package traffic

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/maansaake/arbiter/pkg/module"
	"github.com/maansaake/arbiter/pkg/report"
	"github.com/trebent/zerologr"
)

var (
	reporter report.Reporter //nolint:gochecknoglobals // package-level state for traffic scheduler

	workloads []*workload //nolint:gochecknoglobals // package-level state for traffic scheduler

	// Stop stuff.
	stopChan chan *workload //nolint:gochecknoglobals // package-level state for traffic scheduler

	ErrNoOpsToSchedule = errors.New("there were no operations to schedule")
	ErrZeroRate        = errors.New("operation has a zero rate")
	ErrCleanupTimeout  = errors.New("cleanup timed out")
	ErrRateIssue       = errors.New("rate issue")

	SampleTolerancePerc = 0.05 //nolint:gochecknoglobals // exported config var for tests
)

const (
	defaultMaxWorkers            = 10
	defaultSampleIntervalSeconds = 10

	cleanupTimeout = 5 * time.Second

	minRateForDefaultSample = 30
)

// Run traffic for the input modules using their exposed operations. Traffic
// generation will make operation calls at the specified rates and report
// problems to the reporter. Run() is asynchronous and returns once the main
// go-routine has been started. Run() will monitor the context's done channel
// and stop gracefully once it's closed.
func Run(ctx context.Context, metadata module.Metadata, r report.Reporter) error {
	zerologr.Info("Running traffic generator")
	// Run initialisation of traffic synchronously
	reporter = r

	workloads = make([]*workload, 0, len(metadata))
	for _, meta := range metadata {
		for _, op := range meta.Ops() {
			if op.Disabled {
				zerologr.Info("Skipping disabled operation", "mod", meta.Name(), "op", op.Name)
				continue
			}

			if op.Rate == 0 {
				return fmt.Errorf("%w: %s", ErrZeroRate, op.Name)
			}

			workloads = append(workloads, &workload{
				statLock: &sync.Mutex{},
				mod:      meta.Name(),
				op:       op,
			})
		}
	}

	if len(workloads) == 0 {
		return ErrNoOpsToSchedule
	}

	// Create stop channel that workloads will report to when stopping.
	stopChan = make(chan *workload, len(workloads))

	// Run the workload in a separate go-routine, runs until context is done
	for _, workload := range workloads {
		go workload.run(ctx)
	}

	return nil
}

func Stop() error {
	zerologr.Info("Stopping traffic generator", "workload_count", len(workloads))

	stopCount := 0
	for {
		select {
		case <-time.After(cleanupTimeout):
			zerologr.Error(ErrCleanupTimeout, "Cleanup timed out after "+cleanupTimeout.String())
			return ErrCleanupTimeout
		case workload := <-stopChan:
			zerologr.Info("Workload stopped", "mod", workload.mod, "op", workload.op.Name)
			stopCount++
			if stopCount == len(workloads) {
				zerologr.Info("All workloads have stopped")
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
