package yaml

import (
	"errors"
	"testing"
	"time"

	"tres-bon.se/arbiter/pkg/module/op"
	"tres-bon.se/arbiter/pkg/report"
)

func TestAddOpToModule(t *testing.T) {
	m := newModule()
	m.addOp(&op.Result{Name: "op", Duration: time.Second}, nil)

	v, ok := m.Operations["op"]
	if !ok {
		t.Fatal("'op' expected in map")
	}
	if v.Executions != 1 {
		t.Fatal("expected 1 execution")
	}
	if v.Timing.count != 1 {
		t.Fatal("expected 1 count")
	}
	if v.Timing.Longest != time.Second {
		t.Fatal("expected 1 second")
	}
	if v.Timing.Shortest != time.Second {
		t.Fatal("expected 1 second")
	}
	if v.Timing.Average != time.Second {
		t.Fatal("expected 1 second")
	}
	if v.Timing.total != time.Second {
		t.Fatal("expected 1 second")
	}

	m.addOp(&op.Result{Name: "op", Duration: time.Second}, nil)

	if v.Timing.total != 2*time.Second {
		t.Fatal("expected 2 seconds in total")
	}
	if v.Timing.count != 2 {
		t.Fatal("expected 2 count")
	}
	if v.Timing.Average != time.Second {
		t.Fatal("expected 1 second average")
	}
	if v.Executions != 2 {
		t.Fatal("expected 2 execs")
	}

	m.addOp(&op.Result{Name: "op", Duration: time.Second}, errors.New("error"))
	if v.Executions != 3 {
		t.Fatal("expected 3 execs")
	}
	if v.Timing.count != 2 {
		t.Fatal("expected 2 count")
	}

	m.addOp(&op.Result{Name: "op", Duration: 4 * time.Second}, nil)
	if v.Executions != 4 {
		t.Fatal("expected 4 execs")
	}
	if v.Timing.count != 3 {
		t.Fatal("expected 3 count")
	}
	if v.Timing.Longest != 4*time.Second {
		t.Fatal("expected longest 4 seconds")
	}
	if v.Timing.Average != 2*time.Second {
		t.Fatal("expected average 2 seconds")
	}
	if v.Timing.Shortest != time.Second {
		t.Fatal("expected shortest 1 second")
	}
}

func TestAddLogErrToModule(t *testing.T) {
	m := newModule()
	m.addLogErr(errors.New("log error"))

	if len(m.Log.errs) != 1 {
		t.Fatal("should have been 1 log error")
	}

	m.addLogErr(errors.New("log error"))

	if len(m.Log.errs) != 2 {
		t.Fatal("should have been 2 log errors")
	}
}

func TestAddLogTriggerToModule(t *testing.T) {
	m := newModule()
	m.addLogTrigger(&report.TriggerReport[string]{
		Timestamp: time.Now(),
		Type:      "raise",
		Value:     "some log message",
	})

	if len(m.Log.triggers) != 1 {
		t.Fatal("should have been 1 log error")
	}

	m.addLogTrigger(&report.TriggerReport[string]{
		Timestamp: time.Now(),
		Type:      "raise",
		Value:     "some log message",
	})

	if len(m.Log.triggers) != 2 {
		t.Fatal("should have been 2 log errors")
	}
}

func TestAddCPUToModule(t *testing.T) {
	m := newModule()
	m.addCPU(12.12)

	if m.CPU.readings != 1 {
		t.Fatal("should have been 1")
	}
	if m.CPU.total != 12.12 {
		t.Fatal("should have been 12.12")
	}
	if m.CPU.High != 12.12 {
		t.Fatal("should have been 12.12")
	}
	if m.CPU.Low != 12.12 {
		t.Fatal("should have been 12.12")
	}
	if m.CPU.Average != 12.12 {
		t.Fatal("should have been 12.12")
	}

	m.addCPU(12.10)

	if m.CPU.Average != 12.11 {
		t.Fatal("should have been 12.11")
	}
	if m.CPU.readings != 2 {
		t.Fatal("should have been 2")
	}
	if m.CPU.Low != 12.10 {
		t.Fatal("should have been 12.10")
	}
	if m.CPU.High != 12.12 {
		t.Fatal("should have been 12.12")
	}
}
