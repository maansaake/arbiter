package traffictest

import (
	"context"
	"errors"
	"sync"
	"testing"

	"tres-bon.se/arbiter/pkg/module"
	"tres-bon.se/arbiter/pkg/module/op"
	"tres-bon.se/arbiter/pkg/report"
	"tres-bon.se/arbiter/pkg/traffic"
	log "tres-bon.se/arbiter/pkg/zerologr"
)

func TestRunAndAwaitStop(t *testing.T) {
	log.VFieldName = "v"
	log.SetLogger(log.New(&log.Opts{
		Console: true,
		V:       100,
	}))

	opWg := sync.WaitGroup{}
	opWg.Add(2)

	testmod := newTestModule()
	testmod.(*testmodule).ops = op.Ops{
		{
			Name: "test",
			Rate: 6000,
			Do: func() (op.Result, error) {
				opWg.Done()
				log.Info("doing OP")
				return op.Result{}, nil
			},
		},
	}
	ctx, cancel := context.WithCancel(context.Background())
	err := traffic.Run(ctx, module.Modules{testmod}, &report.YAMLReporter{})
	if err != nil {
		t.Fatal(err)
	}
	log.Info("started traffic")

	opWg.Wait()

	cancel()
	traffic.AwaitStop()
}

func TestRunNoOps(t *testing.T) {
	err := traffic.Run(context.TODO(), module.Modules{newTestModule()}, nil)
	if err != nil && !errors.Is(err, traffic.ErrNoOpsToSchedule) {
		t.Fatal("unexpected error type")
	}
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRunZeroRate(t *testing.T) {
	mod := newTestModule()
	mod.(*testmodule).ops = op.Ops{
		{
			Name: "test",
			Rate: 0,
		},
	}
	err := traffic.Run(context.TODO(), module.Modules{mod}, nil)
	if err != nil && !errors.Is(err, traffic.ErrZeroRate) {
		t.Fatal("unexpected error type")
	}
	if err == nil {
		t.Fatal("expected error")
	}
}
