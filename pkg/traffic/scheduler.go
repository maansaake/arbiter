package traffic

import (
	"context"
	"errors"
	"fmt"
	"time"

	"tres-bon.se/arbiter/pkg/module"
	"tres-bon.se/arbiter/pkg/report"
	"tres-bon.se/arbiter/pkg/subcommand"
	log "tres-bon.se/arbiter/pkg/zerologr"
)

type workload struct {
	mod    string
	op     *module.Op
	ticker *time.Ticker
}

var (
	reporter report.Reporter

	workloads      []*workload
	stop           chan string
	cleanupTimeout = 5 * time.Second
)

var (
	ErrNoOpsToSchedule = errors.New("there were no operations to schedule")
	ErrZeroRate        = errors.New("operation has a zero rate")
)

// Runs traffic for the input modules using their exposed operations. Traffic
// generation will make operation calls at the specified rates and report
// problems to the reporter. Run() is asynchronous and returns once the main
// go-routine has been started. Run() will monitor the context's done channel
// and stop gracefully once it's closed.
func Run(ctx context.Context, metadata subcommand.Metadata, r report.Reporter) error {
	log.Info("running traffic generator")
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
				//nolint:gosec
				ticker: time.NewTicker(time.Minute / time.Duration(op.Rate)),
				op:     op,
			})
		}
	}

	if len(workloads) == 0 {
		return ErrNoOpsToSchedule
	}

	// Start traffic generation in separate go-routine, run until context is done
	// Create stop channel that workloads will report to when stopping.
	stop = make(chan string, len(workloads))
	run(ctx)

	return nil
}

func Stop() {
	log.Info("cleaning up the traffic generator")
	stopCount := 0
	for {
		select {
		case <-time.After(cleanupTimeout):
			log.Error(errors.New("failed to clean up in time"), "timeout", cleanupTimeout)
			return
		case name := <-stop:
			log.Info("workload stopped", "name", name)
			stopCount++
			if stopCount == len(workloads) {
				log.Info("all workloads have stopped")
				return
			}
		}
	}
}

func run(ctx context.Context) {
	for _, workload := range workloads {
		// Index passed to allow easy self deletion.
		go handleWorkload(ctx, workload)
	}
}

func handleWorkload(ctx context.Context, workload *workload) {
	for {
		select {
		case t := <-workload.ticker.C:
			log.V(100).Info("tick", "mod", workload.mod, "op", workload.op.Name, "time", t)

			// Run Op and calculate duration.
			start := time.Now()
			res, err := workload.op.Do()

			if res.Duration == 0 {
				res.Duration = time.Since(start)
			}

			// Report to reporter
			reporter.Op(workload.mod, workload.op.Name, &res, err)
		case <-ctx.Done():
			log.Info("stopping workload", "mod", workload.mod, "op", workload.op.Name)
			workload.ticker.Stop()
			stop <- fmt.Sprintf("%s: %s", workload.mod, workload.op.Name)
			return
		}
	}
}
