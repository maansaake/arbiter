package traffic

import (
	"context"
	"math"
	"sync"
	"time"

	"github.com/maansaake/arbiter/pkg/module"
	"github.com/trebent/zerologr"
)

type workload struct {
	mod string
	op  *module.Op

	workers []*worker

	statLock *sync.Mutex
	calls    float64
	totalDur time.Duration
}

const workloadVerboseLogLevel = 100

// run runs the workload, which in turn spawns workers to do the actual invocations. The workload
// will monitor call-rates and scale the number of workers as needed.
func (w *workload) run(ctx context.Context) {
	zerologr.Info("Starting workload", "mod", w.mod, "op", w.op.Name, "rate", w.op.Rate)

	// All workload start with exactly one worker. After the first sampling
	// period, this may be increased.
	w.workers = make([]*worker, 0, 1)
	w.addWorker(ctx)
	w.calls = 0

	samplingInterval := getSampleInterval(w.op)
	expectedCalls := float64(samplingInterval) / float64(time.Minute) * float64(w.op.Rate)
	zerologr.Info(
		"Setting sampling interval",
		"mod",
		w.mod,
		"op",
		w.op.Name,
		"sampling_interval_ms",
		samplingInterval.Milliseconds(),
		"expected_calls",
		expectedCalls,
	)
	rateCheckTicker := time.NewTicker(samplingInterval)

	for {
		select {
		case <-ctx.Done():
			zerologr.Info("Context closed, stopping workload", "mod", w.mod, "op", w.op.Name)
			rateCheckTicker.Stop()

			for i, worker := range w.workers {
				zerologr.Info("Awaiting worker", "mod", w.mod, "op", w.op.Name, "worker", i)
				// Await the stop of each worker
				<-worker.done
				zerologr.Info("Worker stopped", "mod", w.mod, "op", w.op.Name, "worker", i)
			}

			stopChan <- w
			return
		case <-rateCheckTicker.C:
			zerologr.Info(
				"Running rate check",
				"mod",
				w.mod,
				"op",
				w.op.Name,
				"calls",
				w.calls,
				"expected_calls",
				expectedCalls,
			)

			if w.calls > 0 {
				avgUs := (w.totalDur / time.Duration(w.calls)).Microseconds()
				zerologr.Info("Average exec time", "mod", w.mod, "op", w.op.Name, "avg_µs", avgUs)

				w.scale(ctx)
			}

			// Reset check # of executions next time the checker is run.
			w.withStatLock(func() {
				w.calls = 0
				w.totalDur = 0
			})
		}
	}
}

// withStatLock calls the input function after obtaining the statLock mutex first.
func (w *workload) withStatLock(f func()) {
	w.statLock.Lock()
	f()
	w.statLock.Unlock()
}

// scale checks if the current worker count of the workload is enough to satisfy the desired rate.
// If not, more workers are added. The ticker interval of all workers is reset to match the new
// rate per worker when this is called.
func (w *workload) scale(ctx context.Context) {
	requiredWorkers := w.getWorkerCount()
	zerologr.Info("Required worker count", "required_workers", requiredWorkers, "mod", w.mod, "op", w.op.Name)

	if int(requiredWorkers) > len(w.workers) {
		zerologr.Info("Adding workers", "count", int(requiredWorkers)-len(w.workers), "mod", w.mod, "op", w.op.Name)
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

// workerTickerInterval works out the ticker interval for each worker.
// Ex: 60000ms / 60 = 1000ms ticker interval.
func (w *workload) workerTickerInterval(workerCount float64) time.Duration {
	return time.Duration(
		float64(1*time.Minute) / w.ratePerWorker(workerCount),
	)
}

// ratePerWorker returns the rate per worker, based on the desired rate and worker count.
func (w *workload) ratePerWorker(workerCount float64) float64 {
	return float64(w.op.Rate) / workerCount
}

// getWorkerCount yields a number either below or above 1, if below, 1 worker is enough.
// If above, at least 2 workers are needed to satisfy the rate. Basically, if
// the maximum rate of 1 worker is much higher than the target rate, a number
// below zero is produced, indicating no more than 1 worker is needed. Returns
// a maximum of 50 workers.
func (w *workload) getWorkerCount() float64 {
	return math.Min(math.Ceil(float64(w.op.Rate)/w.getMaxRate()), defaultMaxWorkers)
}

// getMaxRate returns the maximum rate of a single worker, derived from the actual average
// execution time of the operation.
func (w *workload) getMaxRate() float64 {
	return float64(1*time.Minute) / float64(w.getAverageDuration())
}

// getAverageDuration returns the average execution time of the operation, derived from the
// total duration and number of calls.
func (w *workload) getAverageDuration() time.Duration {
	return w.totalDur / time.Duration(w.calls)
}

func (w *workload) addWorker(ctx context.Context) {
	zerologr.Info("Adding worker", "mod", w.mod, "op", w.op.Name)

	worker := &worker{
		parent: w,
		//nolint:gosec // rate is validated to be non-zero before scheduling
		ticker: time.NewTicker(time.Minute / time.Duration(w.op.Rate)),
	}

	w.workers = append(w.workers, worker)
	go worker.run(ctx)
}

// doOp executes the workload operation and reports the result to the reporter. It also updates
// the total duration and call count for the workload, which are used to calculate the average execution time.
func (w *workload) doOp() {
	zerologr.V(workloadVerboseLogLevel).Info("Triggering workload op", "mod", w.mod, "op", w.op.Name)

	start := time.Now()
	res, err := w.op.Do()
	zerologr.V(workloadVerboseLogLevel).Info("Ran op", "mod", w.mod, "op", w.op.Name)

	if res.Duration == 0 {
		res.Duration = time.Since(start)
	}

	// Increase invocation counter and total duration to calculate average
	// execution time.
	w.withStatLock(func() {
		w.calls++
		w.totalDur += res.Duration
	})
	zerologr.V(workloadVerboseLogLevel).Info("Reporting", "mod", w.mod, "op", w.op.Name)

	reporter.Op(w.mod, w.op.Name, &res, err)

	zerologr.V(workloadVerboseLogLevel).
		Info("Trigger done", "mod", w.mod, "op", w.op.Name, "duration_µs", time.Since(start).Microseconds())
}
