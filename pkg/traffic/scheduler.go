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

// Information about an operation to be run.
type (
	workload struct {
		mod string
		op  *module.Op

		workers []*worker
		// Combined rate of all workers, should match op.Rate.
		compoundRate uint
		rateCheck    *time.Ticker
	}
	worker struct {
		done       chan bool
		parent     *workload
		targetRate uint
		ticker     *time.Ticker
	}
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

	Samples             = 100
	SampleTolerancePerc = 0.05
)

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
	zerologr.Info("awaiting cleaning up of the traffic generator", "workload_count", len(workloads))
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

func (w *workload) run(ctx context.Context) {
	zerologr.Info("starting workload", "mod", w.mod, "op", w.op.Name, "rate", w.op.Rate)

	w.compoundRate = 0
	// Set check duration depending on the operation rate, ensure
	// (theoretically) that <Samples> operations can run before checking.
	// This rate is optimistic, and does not account for the time it takes
	// to execute the operation. Nor do the workers initially. The first sample
	// that takes place will determine if the rate of the worker needs to be
	// increased or if a new worker needs to be added.
	w.rateCheck = time.NewTicker(getSampleInterval(w.op.Rate))

	go func() {
		for {
			select {
			case <-ctx.Done():
				zerologr.Info("context closed, stopping workload", "mod", w.mod, "op", w.op.Name)
				w.rateCheck.Stop()

				for i, worker := range w.workers {
					zerologr.Info("awaiting worker", "mod", w.mod, "op", w.op.Name, "worker", i)
					// Await the stop of each worker
					<-worker.done
					zerologr.Info("worker stopped", "mod", w.mod, "op", w.op.Name, "worker", i)
				}

				stop <- w
				return
			case <-w.rateCheck.C:
				zerologr.Info("running rate check", "mod", w.mod, "op", w.op.Name)

				// Reset compound rate to check # of executions next time the checker
				// is run.
				w.compoundRate = 0
			}
		}
	}()

	// All workload start with exactly one worker. After the initial immediate
	// tick, the worker will determine if the parent should be told to increase
	// the worker count. This depends on both the time it takes to execute the
	// operation, and the operation rate. For example, a rate of 60.000 op/min,
	// and a 1ms operation execution time will lead to an additional worker to
	// ensure that the operation is executed at the correct rate.
	w.workers = make([]*worker, 0, 1)
	w.addWorker(ctx)
}

func (w *workload) addWorker(ctx context.Context) {
	zerologr.Info("adding worker", "mod", w.mod, "op", w.op.Name)

	worker := &worker{
		parent:     w,
		targetRate: w.op.Rate,
	}

	w.workers = append(w.workers, worker)
	go worker.run(ctx)
}

func (w *workload) doOp() {
	zerologr.V(100).Info("triggering workload op", "mod", w.mod, "op", w.op.Name)
	w.compoundRate++

	start := time.Now()
	res, err := w.op.Do()

	if res.Duration == 0 {
		res.Duration = time.Since(start)
	}
	reporter.Op(w.mod, w.op.Name, &res, err)
}

func (worker *worker) run(ctx context.Context) {
	zerologr.Info("starting worker", "mod", worker.parent.mod, "op", worker.parent.op.Name)
	//nolint:gosec
	worker.ticker = time.NewTicker(getTargetInterval(worker.targetRate))
	worker.done = make(chan bool, 1)

	for {
		select {
		case <-ctx.Done():
			zerologr.Info("context closed, stopping worker", "mod", worker.parent.mod, "op", worker.parent.op.Name)
			worker.ticker.Stop()

			worker.done <- true
			return
		case t := <-worker.ticker.C:
			zerologr.V(100).Info("worker ticker tick", "time", t, "mod", worker.parent.mod, "op", worker.parent.op.Name)
			worker.parent.doOp()
		}
	}
}

func (worker *worker) reset(newTargetRate uint) {
	zerologr.V(100).Info("resetting worker ticker", "mod", worker.parent.mod, "op", worker.parent.op.Name, "new_target_rate", newTargetRate)
	worker.ticker.Reset(getTargetInterval(newTargetRate))
}

func getTargetInterval(ratePerMinute uint) time.Duration {
	//nolint:gosec
	return time.Minute / time.Duration(ratePerMinute)
}

func getSampleInterval(ratePerMinute uint) time.Duration {
	//nolint:gosec
	return getTargetInterval(ratePerMinute) * time.Duration(Samples)
}
