package traffic

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/maansaake/arbiter/pkg/module"
	modulemock "github.com/maansaake/arbiter/pkg/module/mock"
	reportmock "github.com/maansaake/arbiter/pkg/report/mock"
	log "github.com/trebent/zerologr"
)

func TestRunAndAwaitStop(t *testing.T) {
	opWg := sync.WaitGroup{}
	opWg.Add(2)

	mod := modulemock.NewMock()
	mod.SetOps = module.Ops{
		{
			Name: "test",
			Rate: 60000,
			Do: func() (module.Result, error) {
				opWg.Done()
				log.Info("Doing OP")
				return module.Result{}, nil
			},
		},
	}
	ctx, cancel := context.WithCancel(context.Background())
	err := Run(ctx, []*module.Meta{{Module: mod}}, reportmock.NewMock())
	if err != nil {
		t.Fatal(err)
	}
	log.Info("Started traffic")

	opWg.Wait()

	cancel()
	err = Stop()
	if err != nil {
		t.Fatal(err)
	}
}

func TestRunNoOps(t *testing.T) {
	err := Run(context.TODO(), []*module.Meta{{Module: modulemock.NewMock()}}, nil)
	if err != nil && !errors.Is(err, ErrNoOpsToSchedule) {
		t.Fatal("unexpected error type")
	}
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRunZeroRate(t *testing.T) {
	mod := modulemock.NewMock()
	mod.SetOps = module.Ops{
		{
			Name: "test",
			Rate: 0,
		},
	}
	err := Run(context.TODO(), []*module.Meta{{Module: mod}}, nil)
	if err != nil && !errors.Is(err, ErrZeroRate) {
		t.Fatal("unexpected error type")
	}
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestReportOpToReporter(t *testing.T) {
	wg := sync.WaitGroup{}
	wg.Add(1)

	mod := modulemock.NewMock()
	mod.SetOps = module.Ops{
		{
			Name: "test",
			Rate: 6000,
			Do: func() (module.Result, error) {
				defer wg.Done()
				return module.Result{}, nil
			},
		},
	}

	reporter := reportmock.NewMock()
	ctx, cancel := context.WithCancel(context.Background())
	if err := Run(ctx, []*module.Meta{{Module: mod}}, reporter); err != nil {
		t.Fatal(err)
	}

	wg.Wait()
	cancel()

	// This should ensure the reporter mock has received the Op report.
	Stop()

	if reporter.OpResults[0].Duration == 0 {
		t.Fatal("duration was not reported to reporter")
	}
}

func TestReportOpDurationOverrideToReporter(t *testing.T) {
	wg := sync.WaitGroup{}
	wg.Add(1)

	mod := modulemock.NewMock()
	mod.SetOps = module.Ops{
		{
			Name: "test",
			Rate: 6000,
			Do: func() (module.Result, error) {
				defer wg.Done()
				return module.Result{Duration: 12 * time.Millisecond}, nil
			},
		},
	}

	reporter := reportmock.NewMock()
	ctx, cancel := context.WithCancel(context.Background())
	if err := Run(ctx, []*module.Meta{{Module: mod}}, reporter); err != nil {
		t.Fatal(err)
	}

	wg.Wait()
	cancel()

	// This should ensure the reporter mock has received the Op report.
	Stop()

	if reporter.OpResults[0].Duration != 12*time.Millisecond {
		t.Fatal("duration override was not used")
	}
}

func TestReportOpErr(t *testing.T) {
	wg := sync.WaitGroup{}
	wg.Add(1)

	mod := modulemock.NewMock()
	mod.SetOps = module.Ops{
		{
			Name: "test",
			Rate: 6000,
			Do: func() (module.Result, error) {
				defer wg.Done()
				return module.Result{Duration: 12 * time.Millisecond}, errors.New("some error")
			},
		},
	}

	reporter := reportmock.NewMock()
	ctx, cancel := context.WithCancel(context.Background())
	if err := Run(ctx, []*module.Meta{{Module: mod}}, reporter); err != nil {
		t.Fatal(err)
	}

	wg.Wait()
	cancel()

	// This should ensure the reporter mock has received the Op report.
	Stop()

	if len(reporter.OpResults) != 0 {
		t.Fatal("unexpected op results found")
	}
	if len(reporter.OpErrors) != 1 {
		t.Fatal("unexpected number of op errors")
	}
}

func TestWorkerTickerInterval(t *testing.T) {
	workload := &workload{
		op:       &module.Op{Rate: 1000},
		calls:    999,
		totalDur: 999 * time.Millisecond,
	}
	// 1ms / op
	maxRate := workload.getMaxRate()

	if maxRate != float64(1*time.Minute)/float64(workload.getAverageDuration()) {
		t.Fatal("unexpected max rate")
	}
	t.Log("max rate:", maxRate)

	workerCount := workload.getWorkerCount()
	if workerCount != 1 {
		t.Fatal("unexpected worker count")
	}

	ratePerWorker := workload.ratePerWorker(workerCount)
	if ratePerWorker != 1000 {
		t.Fatal("unexpected worker rate")
	}

	workerTickerInterval := workload.workerTickerInterval(workerCount)
	t.Log(workerTickerInterval)
	if workerTickerInterval != 1*time.Minute/time.Duration(workload.op.Rate) {
		t.Fatal("unexpected worker ticker interval")
	}
}
