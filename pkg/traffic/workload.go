package traffic

import (
	"context"
	"math"
	"time"

	"tres-bon.se/arbiter/pkg/module"
	"tres-bon.se/arbiter/pkg/zerologr"
)

type workload struct {
	mod string
	op  *module.Op

	workers []*worker

	// Combined rate of all workers, should match op.Rate.
	calls     float64
	totalDur  time.Duration
	rateCheck *time.Ticker
}

func (w *workload) run(ctx context.Context) {
	zerologr.Info("starting workload", "mod", w.mod, "op", w.op.Name, "rate", w.op.Rate)

	go func() {
		w.calls = 0
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
				zerologr.Info("running rate check", "mod", w.mod, "op", w.op.Name, "calls", w.calls, "expected_calls", expectedCalls, "avg_µs", (w.totalDur / time.Duration(w.calls)).Microseconds())

				w.scale(ctx)

				// Reset compound rate to check # of executions next time the checker
				// is run.
				w.totalDur = 0
				w.calls = 0
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
	zerologr.Info("required worker count", "required_workers", requiredWorkers, "mod", w.mod, "op", w.op.Name)
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

// Work out the ticker interval for each worker.
// Ex: 60000ms / 60 = 1000ms ticker interval
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
	return w.totalDur / time.Duration(w.calls)
}

func (w *workload) addWorker(ctx context.Context) {
	zerologr.Info("adding worker", "mod", w.mod, "op", w.op.Name)

	worker := &worker{
		parent: w,
		//nolint:gosec
		ticker: time.NewTicker(time.Minute / time.Duration(w.op.Rate)),
	}

	w.workers = append(w.workers, worker)
	go worker.run(ctx)
}

func (w *workload) doOp() {
	zerologr.V(100).Info("triggering workload op", "mod", w.mod, "op", w.op.Name)

	start := time.Now()
	res, err := w.op.Do()
	zerologr.V(100).Info("ran op", "mod", w.mod, "op", w.op.Name)

	if res.Duration == 0 {
		res.Duration = time.Since(start)
	}

	// Increase invocation counter and total duration to calculate average
	// execution time.
	w.calls++
	w.totalDur += res.Duration
	zerologr.V(100).Info("reporting", "mod", w.mod, "op", w.op.Name)

	reporter.Op(w.mod, w.op.Name, &res, err)

	zerologr.V(100).Info("trigger done", "mod", w.mod, "op", w.op.Name, "duration_µs", time.Since(start).Microseconds())
}
