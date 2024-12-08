package monitor

import (
	"context"
	"errors"
	"testing"
	"time"

	"tres-bon.se/arbiter/pkg/monitor/cpu"
	"tres-bon.se/arbiter/pkg/monitor/log"
	"tres-bon.se/arbiter/pkg/monitor/memory"
	"tres-bon.se/arbiter/pkg/monitor/metric"
)

func TestStartMonitorCPU(t *testing.T) {
	monitor := New()
	monitorInterval = 1 * time.Microsecond

	mock := cpu.NewCPUMonitorMock(0, nil)
	ctx, cancel := context.WithCancel(context.Background())
	mock.SetOnRead(func() {
		cancel()
	})
	monitor.cpuMonitors["bla"] = mock

	err := monitor.Start(ctx)
	if err != nil {
		t.Fatalf("error starting monitor %v", err)
	}

	<-ctx.Done()
}

func TestStartMonitorMemory(t *testing.T) {
	monitor := New()
	monitorInterval = 1 * time.Microsecond

	mock := memory.NewMemoryMock(0, 0, nil)
	ctx, cancel := context.WithCancel(context.Background())
	mock.SetOnRead(func() {
		cancel()
	})
	monitor.memMonitors["bla"] = mock

	err := monitor.Start(ctx)
	if err != nil {
		t.Fatalf("error starting monitor %v", err)
	}

	<-ctx.Done()
}

func TestStartMonitorMetric(t *testing.T) {
	monitor := New()
	monitorInterval = 1 * time.Microsecond

	mock := metric.NewMetricMock(nil, nil)
	ctx, cancel := context.WithCancel(context.Background())
	mock.SetOnRead(func() {
		cancel()
	})
	monitor.metricMonitors["bla"] = mock

	err := monitor.Start(ctx)
	if err != nil {
		t.Fatalf("error starting monitor %v", err)
	}

	<-ctx.Done()
}

func TestStartMonitorLog(t *testing.T) {
	monitor := New()
	monitorInterval = 1 * time.Microsecond

	mock := log.NewLogMock(nil)
	ctx, cancel := context.WithCancel(context.Background())
	mock.SetOnRead(func() {
		cancel()
	})
	monitor.logMonitors["bla"] = mock

	err := monitor.Start(ctx)
	if err != nil {
		t.Fatalf("error starting monitor %v", err)
	}

	<-ctx.Done()
}

func TestStartMonitorLogErr(t *testing.T) {
	monitor := New()
	monitorInterval = 1 * time.Microsecond

	mock := log.NewLogMock(errors.New("bla"))
	ctx, cancel := context.WithCancel(context.Background())
	monitor.logMonitors["bla"] = mock

	err := monitor.Start(ctx)
	if err == nil {
		t.Fatalf("monitor start should have thrown an error %v", err)
	}

	cancel()
}
