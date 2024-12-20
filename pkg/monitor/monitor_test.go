package monitor

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"tres-bon.se/arbiter/pkg/monitor/cpu"
	"tres-bon.se/arbiter/pkg/monitor/log"
	"tres-bon.se/arbiter/pkg/monitor/memory"
	"tres-bon.se/arbiter/pkg/monitor/metric"
	"tres-bon.se/arbiter/pkg/monitor/trigger"
)

func TestStartMonitorCPU(t *testing.T) {
	monitor := New()
	monitor.ExternalPrometheus = true
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
	err = monitor.Stop()
	if err != nil {
		t.Fatal("failed to stop monitor", err)
	}
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
	err = monitor.Stop()
	if err != nil {
		t.Fatal("failed to stop monitor")
	}
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
	err = monitor.Stop()
	if err != nil {
		t.Fatal("failed to stop monitor", err)
	}
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
	err = monitor.Stop()
	if err != nil {
		t.Fatal("failed to stop monitor", err)
	}
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

func TestAdd(t *testing.T) {
	pidOpt := DefaultOpt()
	pidOpt.Name = "pid"
	pidOpt.PID = os.Getpid()
	m := New(pidOpt)
	m.ExternalPrometheus = true

	if _, ok := m.cpuMonitors["pid"]; !ok {
		t.Fatal("expected 1 CPU monitor")
	}
	if _, ok := m.memMonitors["pid"]; !ok {
		t.Fatal("expected 1 memory monitor")
	}

	metricOpt := DefaultOpt()
	metricOpt.Name = "metric"
	metricOpt.MetricEndpoint = "not none"
	metricOpt.MetricTriggers["metric"] = []trigger.Trigger[float64]{}
	m.Add(metricOpt)

	if _, ok := m.metricMonitors["metric"]; !ok {
		t.Fatal("expected 1 metric monitor")
	}

	logOpt := DefaultOpt()
	logOpt.Name = "log"
	logOpt.LogFile = "logfilepath"
	m.Add(logOpt)

	if _, ok := m.logMonitors["log"]; !ok {
		t.Fatal("expected 1 log monitor")
	}
}

func TestOpt(t *testing.T) {
	opt := DefaultOpt()
	opt.Name = "name"

	opt.CPUTriggerFromCmdline("ABOVE;12")
	opt.VMSTriggerFromCmdline("ABOVE;12")
	opt.RSSTriggerFromCmdline("ABOVE;12")
	opt.MetricTriggerFromCmdline("metric;ABOVE;12")
	opt.LogFileTriggerFromCmdline("EQUAL;phrase")

	if len(opt.CPUTriggers) != 1 {
		t.Fatal("CPU trigger expected to be added")
	}
	if len(opt.VMSTriggers) != 1 {
		t.Fatal("VMS trigger expected to be added")
	}
	if len(opt.RSSTriggers) != 1 {
		t.Fatal("RSS trigger expected to be added")
	}
	if len(opt.MetricTriggers) != 1 {
		t.Fatal("Metric trigger expected to be added")
	}
	if len(opt.LogTriggers) != 1 {
		t.Fatal("Log trigger expected to be added")
	}
}
