package monitor

import (
	"context"
	"testing"
	"time"

	"tres-bon.se/arbiter/pkg/monitor/cpu"
	"tres-bon.se/arbiter/pkg/monitor/memory"
)

func TestStartMonitorCPU(t *testing.T) {
	monitor := &Monitor{}
	monitorInterval = 1 * time.Microsecond

	mock := cpu.NewCPUMonitorMock(0, nil)
	ctx, cancel := context.WithCancel(context.Background())
	mock.SetOnRead(func() {
		cancel()
	})
	monitor.CPU = mock

	err := monitor.Start(ctx)
	if err != nil {
		t.Fatalf("error starting monitor %v", err)
	}

	<-ctx.Done()
}

func TestStartMonitorMemory(t *testing.T) {
	monitor := &Monitor{}
	monitorInterval = 1 * time.Microsecond

	mock := memory.NewMemoryMock(0, 0, nil)
	ctx, cancel := context.WithCancel(context.Background())
	mock.SetOnRead(func() {
		cancel()
	})
	monitor.Memory = mock

	err := monitor.Start(ctx)
	if err != nil {
		t.Fatalf("error starting monitor %v", err)
	}

	<-ctx.Done()
}
