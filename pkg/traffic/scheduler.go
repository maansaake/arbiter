package traffic

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"tres-bon.se/arbiter/pkg/module/op"
	"tres-bon.se/arbiter/pkg/report"
	"tres-bon.se/arbiter/pkg/subcommand"
	log "tres-bon.se/arbiter/pkg/zerologr"
)

type workload struct {
	ticker *time.Ticker
	op     *op.Op
}

var (
	reporter report.Reporter

	workloads []*workload
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
func Run(ctx context.Context, meta []*subcommand.ModuleMeta, r report.Reporter) error {
	log.Info("running traffic generator")
	// Run initialisation of traffic synchronously
	reporter = r

	ops := make([]*op.Op, 0, 5)
	for _, m := range meta {
		for _, op := range m.Ops() {
			if op.Disabled {
				continue
			}

			if op.Rate == 0 {
				return fmt.Errorf("%w: %s", ErrZeroRate, op.Name)
			}

			ops = append(ops, op)
		}
	}

	if len(ops) == 0 {
		return ErrNoOpsToSchedule
	}

	workloads = make([]*workload, len(ops))
	for i, op := range ops {
		// Tick op.Rate times per minute
		workloads[i] = &workload{
			ticker: time.NewTicker(time.Minute / time.Duration(op.Rate)),
			op:     op,
		}
	}

	// Start traffic generation in separate go-routine, run until context is done
	run(ctx)

	return nil
}

func AwaitStop() {
	log.Info("awaiting cleanup of traffic generator")
	for {
		<-time.After(1 * time.Millisecond)
		// We want there to be only nil pointers, then tickers have been cleaned
		// up.
		if slices.IndexFunc(
			workloads, func(wl *workload) bool { return wl != nil },
		) == -1 {
			return
		}
	}
}

func run(ctx context.Context) {
	for i, workload := range workloads {
		// Index passed to allow easy self deletion.
		go handleWorkload(ctx, i, workload)
	}
}

func handleWorkload(ctx context.Context, index int, workload *workload) {
	for {
		select {
		case <-ctx.Done():
			log.Info("stopping workload for operation", "op", workload.op.Name)
			workload.ticker.Stop()
			workloads[index] = nil
			return
		case t := <-workload.ticker.C:
			log.V(100).Info("tick", "op", workload.op.Name, "time", t)

			// Run Op and calculate duration.
			start := time.Now()
			res, err := workload.op.Do()

			// If override is set by the user module, use the overridden duration.
			if res.DurationOverride != 0 {
				res.Duration = res.DurationOverride
				// Else calculate from start timestamp
			} else {
				res.Duration = time.Since(start)
			}

			// Report to reporter
			reporter.Op(&res, err)
		}
	}
}
