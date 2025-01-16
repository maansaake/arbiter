package traffic

import (
	"context"
	"time"

	"tres-bon.se/arbiter/pkg/zerologr"
)

type worker struct {
	done   chan bool
	parent *workload
	ticker *time.Ticker
}

func (worker *worker) run(ctx context.Context) {
	zerologr.Info("starting worker", "mod", worker.parent.mod, "op", worker.parent.op.Name)
	worker.done = make(chan bool)

	for {
		select {
		case <-ctx.Done():
			zerologr.Info("context closed, stopping worker", "mod", worker.parent.mod, "op", worker.parent.op.Name)
			worker.ticker.Stop()

			close(worker.done)
			return
		case t := <-worker.ticker.C:
			zerologr.V(100).Info("worker tick", "time", t, "mod", worker.parent.mod, "op", worker.parent.op.Name)
			worker.parent.doOp()
		}
	}
}

func (worker *worker) reset(tickerInterval time.Duration) {
	zerologr.Info("resetting worker ticker", "mod", worker.parent.mod, "op", worker.parent.op.Name, "interval_µs", tickerInterval.Microseconds())
	worker.ticker.Reset(tickerInterval)
}
