package traffic

import (
	"context"
	"time"
)

const workerVerboseLogLevel = 100

type worker struct {
	done   chan bool
	parent *workload
	ticker *time.Ticker
}

func (worker *worker) run(ctx context.Context) {
	logger.Info("Starting worker", "mod", worker.parent.mod, "op", worker.parent.op.Name)
	worker.done = make(chan bool)

	for {
		select {
		case <-ctx.Done():
			logger.Info("Context closed, stopping worker", "mod", worker.parent.mod, "op", worker.parent.op.Name)
			worker.ticker.Stop()

			close(worker.done)
			return
		case t := <-worker.ticker.C:
			logger.V(workerVerboseLogLevel).
				Info("Worker tick", "time", t, "mod", worker.parent.mod, "op", worker.parent.op.Name)
			worker.parent.doOp()
		}
	}
}

func (worker *worker) reset(tickerInterval time.Duration) {
	logger.Info(
		"Resetting worker ticker",
		"mod",
		worker.parent.mod,
		"op",
		worker.parent.op.Name,
		"interval_µs",
		tickerInterval.Microseconds(),
	)
	worker.ticker.Reset(tickerInterval)
}
