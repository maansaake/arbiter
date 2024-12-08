package traffic

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"tres-bon.se/arbiter/pkg/module/op"
	testmodule "tres-bon.se/arbiter/pkg/module/test"
	"tres-bon.se/arbiter/pkg/report"
	mockreport "tres-bon.se/arbiter/pkg/report/mock"
	"tres-bon.se/arbiter/pkg/subcommand"
	log "tres-bon.se/arbiter/pkg/zerologr"
)

func TestRunAndAwaitStop(t *testing.T) {
	// log.VFieldName = "v"
	// log.SetLogger(log.New(&log.Opts{
	// 	Console: true,
	// 	V:       100,
	// }))

	opWg := sync.WaitGroup{}
	opWg.Add(2)

	testmod := testmodule.NewTestModule()
	testmod.(*testmodule.TestModule).SetOps = op.Ops{
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
	err := Run(ctx, []*subcommand.Meta{{Module: testmod}}, &report.YAMLReporter{})
	if err != nil {
		t.Fatal(err)
	}
	log.Info("started traffic")

	opWg.Wait()

	cancel()
	AwaitStop()
}

func TestRunNoOps(t *testing.T) {
	err := Run(context.TODO(), []*subcommand.Meta{{Module: testmodule.NewTestModule()}}, nil)
	if err != nil && !errors.Is(err, ErrNoOpsToSchedule) {
		t.Fatal("unexpected error type")
	}
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRunZeroRate(t *testing.T) {
	mod := testmodule.NewTestModule()
	mod.(*testmodule.TestModule).SetOps = op.Ops{
		{
			Name: "test",
			Rate: 0,
		},
	}
	err := Run(context.TODO(), []*subcommand.Meta{{Module: mod}}, nil)
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

	mod := testmodule.NewTestModule()
	mod.(*testmodule.TestModule).SetOps = op.Ops{
		{
			Name: "test",
			Rate: 6000,
			Do: func() (op.Result, error) {
				defer wg.Done()
				return op.Result{}, nil
			},
		},
	}

	reporter := mockreport.NewReporterMock()
	ctx, cancel := context.WithCancel(context.Background())
	if err := Run(ctx, []*subcommand.Meta{{Module: mod}}, reporter); err != nil {
		t.Fatal(err)
	}

	wg.Wait()
	cancel()

	// This should ensure the reporter mock has received the Op report.
	AwaitStop()

	if reporter.(*mockreport.ReporterMock).Results[0].Duration == 0 {
		t.Fatal("duration was not reported to reporter")
	}
}

func TestReportOpDurationOverrideToReporter(t *testing.T) {
	wg := sync.WaitGroup{}
	wg.Add(1)

	mod := testmodule.NewTestModule()
	mod.(*testmodule.TestModule).SetOps = op.Ops{
		{
			Name: "test",
			Rate: 6000,
			Do: func() (op.Result, error) {
				defer wg.Done()
				return op.Result{DurationOverride: 12 * time.Millisecond}, nil
			},
		},
	}

	reporter := mockreport.NewReporterMock()
	ctx, cancel := context.WithCancel(context.Background())
	if err := Run(ctx, []*subcommand.Meta{{Module: mod}}, reporter); err != nil {
		t.Fatal(err)
	}

	wg.Wait()
	cancel()

	// This should ensure the reporter mock has received the Op report.
	AwaitStop()

	if reporter.(*mockreport.ReporterMock).Results[0].Duration != 12*time.Millisecond {
		t.Fatal("duration override was not used")
	}
}
