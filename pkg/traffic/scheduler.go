package traffic

import (
	"context"
	"errors"
	"fmt"
	"math"
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
		rateCount float64
		totalDur  time.Duration
		rateCheck *time.Ticker
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

func (w *workload) run(ctx context.Context) {
	zerologr.Info("starting workload", "mod", w.mod, "op", w.op.Name, "rate", w.op.Rate)

	go func() {
		w.rateCount = 0
		samplingInterval := getSampleInterval(w.op)
		expectedCalls := float64(samplingInterval) / float64(time.Minute) * float64(w.op.Rate)
		zerologr.Info("setting sampling interval", "mod", w.mod, "op", w.op.Name, "sampling_interval_ms", samplingInterval.Milliseconds(), "expected_calls", expectedCalls)
		w.rateCheck = time.NewTicker(samplingInterval)

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
				zerologr.Info("running rate check", "mod", w.mod, "op", w.op.Name, "compound_rate", w.rateCount, "expected_calls", expectedCalls)

				if w.rateOutOfBounds(expectedCalls) {
					zerologr.Info("rate out of bounds", "mod", w.mod, "op", w.op.Name, "average_duration_ms", w.getAverageDuration().Milliseconds())
					w.scale(ctx)
				}

				// Reset compound rate to check # of executions next time the checker
				// is run.
				w.totalDur = 0
				w.rateCount = 0
			}
		}
	}()

	// All workload start with exactly one worker. After the first sampling
	// period, this may be increased.
	w.workers = make([]*worker, 0, 1)
	w.addWorker(ctx)
}

func (w *workload) scale(ctx context.Context) {
	requiredWorkers := w.getWorkerCount()
	zerologr.Info("calculated required worker count", "required_workers", requiredWorkers, "mod", w.mod, "op", w.op.Name)
	if int(requiredWorkers) > len(w.workers) {
		zerologr.Info("adding workers", "count", int(requiredWorkers)-len(w.workers), "mod", w.mod, "op", w.op.Name)
		for i := int(requiredWorkers) - len(w.workers); i > 0; i-- {
			w.addWorker(ctx)
		}
	}

	// Reset the worker tickers to match the new rate. This applies to both
	// scaling up and down.
	for _, worker := range w.workers {
		worker.reset(w.workerTickerInterval(requiredWorkers))
	}
}

func (w *workload) rateOutOfBounds(expectedCalls float64) bool {
	return (w.rateCount < (expectedCalls*(1-SampleTolerancePerc)) || w.rateCount > (expectedCalls*(1+SampleTolerancePerc)))
}

// Work out the ticker interval for each worker, based on the rate and the
// average duration it takes to execute each operation.
// Ex: 60000ms / 60 - 1ms = 999ms ticker interval
func (w *workload) workerTickerInterval(workerCount float64) time.Duration {
	return time.Duration(
		float64(1*time.Minute) / w.ratePerWorker(workerCount),
	)
}

// Returns the rate per worker, based on the desired rate and worker count.
func (w *workload) ratePerWorker(workerCount float64) float64 {
	return float64(w.op.Rate) / workerCount
}

// This yields a number either below or above 1, if below, 1 worker is enough.
// If above, at least 2 workers are needed to satisfy the rate. Basically, if
// the maximum rate of 1 worker is much higher than the target rate, a number
// below zero is produced, indicating no more than 1 worker is needed. Returns
// a maximum of 50 workers.
func (w *workload) getWorkerCount() float64 {
	return math.Min(math.Ceil(float64(w.op.Rate)/w.getMaxRate()), maxWorkers)
}

// Returns the maximum rate of a single worker, derived from the actual average
// execution time of the operation.
func (w *workload) getMaxRate() float64 {
	return float64(1*time.Minute) / float64(w.getAverageDuration())
}

func (w *workload) getAverageDuration() time.Duration {
	return w.totalDur / time.Duration(w.rateCount)
}

func (w *workload) addWorker(ctx context.Context) {
	zerologr.Info("adding worker", "mod", w.mod, "op", w.op.Name)

	worker := &worker{
		parent: w,
		//nolint:gosec
		ticker:     time.NewTicker(time.Minute / time.Duration(w.op.Rate)),
		targetRate: w.op.Rate,
	}

	w.workers = append(w.workers, worker)
	go worker.run(ctx)
}

func (w *workload) doOp() {
	zerologr.V(100).Info("triggering workload op", "mod", w.mod, "op", w.op.Name)

	start := time.Now()
	res, err := w.op.Do()

	if res.Duration == 0 {
		res.Duration = time.Since(start)
	}

	// Increase invocation counter and total duration to calculate average
	// execution time.
	w.rateCount++
	w.totalDur += res.Duration

	reporter.Op(w.mod, w.op.Name, &res, err)
}

func (worker *worker) run(ctx context.Context) {
	zerologr.Info("starting worker", "mod", worker.parent.mod, "op", worker.parent.op.Name)
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

func (worker *worker) reset(tickerInterval time.Duration) {
	zerologr.Info("resetting worker ticker", "mod", worker.parent.mod, "op", worker.parent.op.Name, "interval", tickerInterval)
	worker.ticker.Reset(tickerInterval)
}

func getSampleInterval(op *module.Op) time.Duration {
	if op.Rate < 30 {
		// Minimum 5 samples, this should be a super corner case.
		//nolint:gosec
		return time.Minute / time.Duration(op.Rate) * 5
	}

	return 10 * time.Second
}
