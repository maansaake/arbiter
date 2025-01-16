package traffic

import (
	"context"
	"errors"
	"fmt"
	"time"

	"tres-bon.se/arbiter/pkg/module"
	"tres-bon.se/arbiter/pkg/report"
	"tres-bon.se/arbiter/pkg/subcommand"
	"tres-bon.se/arbiter/pkg/zerologr"
)

var (
	reporter report.Reporter

	workloads []*workload

	// Stop stuff.
	stop           chan *workload
	cleanupTimeout = 5 * time.Second

	ErrNoOpsToSchedule = errors.New("there were no operations to schedule")
	ErrZeroRate        = errors.New("operation has a zero rate")
	ErrCleanupTimeout  = errors.New("cleanup timed out")
	ErrRateIssue       = errors.New("rate issue")

	SampleTolerancePerc = 0.05
)

const maxWorkers = 50

// Runs traffic for the input modules using their exposed operations. Traffic
// generation will make operation calls at the specified rates and report
// problems to the reporter. Run() is asynchronous and returns once the main
// go-routine has been started. Run() will monitor the context's done channel
// and stop gracefully once it's closed.
func Run(ctx context.Context, metadata subcommand.Metadata, r report.Reporter) error {
	zerologr.Info("running traffic generator")
	// Run initialisation of traffic synchronously
	reporter = r

	workloads = make([]*workload, 0, len(metadata)*2)
	for _, meta := range metadata {
		for _, op := range meta.Ops() {
			if op.Disabled {
				continue
			}

			if op.Rate == 0 {
				return fmt.Errorf("%w: %s", ErrZeroRate, op.Name)
			}

			workloads = append(workloads, &workload{
				mod: meta.Name(),
				op:  op,
			})
		}
	}

	if len(workloads) == 0 {
		return ErrNoOpsToSchedule
	}

	// Create stop channel that workloads will report to when stopping.
	stop = make(chan *workload, len(workloads))
	// Start traffic generation in separate go-routine, runs until context is done
	for _, workload := range workloads {
		go workload.run(ctx)
	}

	return nil
}

func AwaitStop() error {
	zerologr.Info("awaiting cleanup of traffic generator", "workload_count", len(workloads))
	stopCount := 0
	for {
		select {
		case <-time.After(cleanupTimeout):
			zerologr.Error(ErrCleanupTimeout, "cleanup timed out", "timeout", cleanupTimeout)
			return ErrCleanupTimeout
		case workload := <-stop:
			zerologr.Info("workload stopped", "mod", workload.mod, "op", workload.op.Name)
			stopCount++
			if stopCount == len(workloads) {
				zerologr.Info("all workloads have stopped")
				return nil
			}
		}
	}
}

func getSampleInterval(op *module.Op) time.Duration {
	if op.Rate < 30 {
		// Minimum 5 samples, this should be a super corner case. Add some time
		// to allow the 5th invocation to fire.
		//nolint:gosec
		return time.Minute/time.Duration(op.Rate)*5 + 250*time.Millisecond
	}

	return 10 * time.Second
}
