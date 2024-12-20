package yaml

import (
	"errors"
	"testing"
	"time"

	"tres-bon.se/arbiter/pkg/module"
	"tres-bon.se/arbiter/pkg/report"
)

func TestAddOpToModule(t *testing.T) {
	m := newModule()
	m.addOp("op", &module.Result{Duration: time.Second}, nil)

	v, ok := m.Operations["op"]
	if !ok {
		t.Fatal("'op' expected in map")
	}
	if v.Executions != 1 {
		t.Fatal("expected 1 execution")
	}
	if v.OK != 1 {
		t.Fatal("expected 1 OK")
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

	m.addOp("op", &module.Result{Duration: time.Second}, nil)

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
	if v.OK != 2 {
		t.Fatal("expected 2 OK")
	}
	if v.NOK != 0 {
		t.Fatal("expected 0 NOK")
	}

	m.addOp("op", &module.Result{Duration: time.Second}, errors.New("error"))
	if v.Executions != 3 {
		t.Fatal("expected 3 execs")
	}
	if v.NOK != 1 {
		t.Fatal("expected 1 NOK")
	}
	if v.Timing.count != 2 {
		t.Fatal("expected 2 count")
	}

	m.addOp("op", &module.Result{Duration: 4 * time.Second}, nil)
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

	if len(m.Log.Errs) != 1 {
		t.Fatal("should have been 1 log error")
	}

	m.addLogErr(errors.New("log error"))

	if len(m.Log.Errs) != 2 {
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

	if len(m.Log.Triggers) != 1 {
		t.Fatal("should have been 1 log triggers")
	}

	m.addLogTrigger(&report.TriggerReport[string]{
		Timestamp: time.Now(),
		Type:      "raise",
		Value:     "some log message",
	})

	if len(m.Log.Triggers) != 2 {
		t.Fatal("should have been 2 log triggers")
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

func TestAddCPUErrToModule(t *testing.T) {
	m := newModule()
	m.addCPUErr(errors.New("some error"))

	if len(m.CPU.Errs) != 1 {
		t.Fatal("should have been 1")
	}
}

func TestAddCPUTriggerToModule(t *testing.T) {
	m := newModule()
	m.addCPUTrigger(&report.TriggerReport[float64]{
		Timestamp: time.Now(),
		Type:      "raise",
		Value:     12.12,
	})

	if len(m.CPU.Triggers) != 1 {
		t.Fatal("should have been 1")
	}
}

func TestAddRSSToModule(t *testing.T) {
	m := newModule()
	m.addRSS(12)

	if m.Mem.RSS.readings != 1 {
		t.Fatal("should have been 1 reading")
	}
	if m.Mem.RSS.High != 12 {
		t.Fatal("high should have been 12")
	}
	if m.Mem.RSS.Low != 12 {
		t.Fatal("low should have been 12")
	}
	if m.Mem.RSS.Average != 12 {
		t.Fatal("average should have been 12")
	}

	m.addRSS(28)
	if m.Mem.RSS.High != 28 {
		t.Fatal("high should have been 28")
	}
	if m.Mem.RSS.Low != 12 {
		t.Fatal("low should have been 12")
	}
	if m.Mem.RSS.Average != 20 {
		t.Fatal("average should have been 20")
	}
	if m.Mem.RSS.readings != 2 {
		t.Fatal("readings should have been 2")
	}
}

func TestAddRSSErrToModule(t *testing.T) {
	m := newModule()
	m.addRSSErr(errors.New("some error"))

	if len(m.Mem.RSS.Errs) != 1 {
		t.Fatal("num errors should have been 1")
	}

	m.addRSSErr(errors.New("some error"))

	if len(m.Mem.RSS.Errs) != 2 {
		t.Fatal("num errors should have been 2")
	}
}

func TestAddRSSTriggerToModule(t *testing.T) {
	m := newModule()
	m.addRSSTrigger(&report.TriggerReport[uint]{
		Timestamp: time.Now(),
		Type:      "raise",
		Value:     12,
	})

	if len(m.Mem.RSS.Triggers) != 1 {
		t.Fatal("num errors should have been 1")
	}

	m.addRSSTrigger(&report.TriggerReport[uint]{
		Timestamp: time.Now(),
		Type:      "raise",
		Value:     12,
	})

	if len(m.Mem.RSS.Triggers) != 2 {
		t.Fatal("num errors should have been 2")
	}
}

func TestAddVMSToModule(t *testing.T) {
	m := newModule()
	m.addVMS(12)

	if m.Mem.VMS.readings != 1 {
		t.Fatal("readings should have been 1")
	}
	if m.Mem.VMS.High != 12 {
		t.Fatal("high should have been 12")
	}
	if m.Mem.VMS.Low != 12 {
		t.Fatal("low should have been 12")
	}
	if m.Mem.VMS.Average != 12 {
		t.Fatal("average should have been 12")
	}

	m.addVMS(10)

	if m.Mem.VMS.High != 12 {
		t.Fatal("high should have been 12")
	}
	if m.Mem.VMS.Low != 10 {
		t.Fatal("low should have been 10")
	}
	if m.Mem.VMS.Average != 11 {
		t.Fatal("average should have been 11")
	}
}

func TestAddVMSErrToModule(t *testing.T) {
	m := newModule()
	m.addVMSErr(errors.New("some error"))

	if len(m.Mem.VMS.Errs) != 1 {
		t.Fatal("num VMS errors should be 1")
	}

	m.addVMSErr(errors.New("some error"))

	if len(m.Mem.VMS.Errs) != 2 {
		t.Fatal("num VMS errors should be 2")
	}
}

func TestAddVMSTriggerToModule(t *testing.T) {
	m := newModule()
	m.addVMSTrigger(&report.TriggerReport[uint]{
		Timestamp: time.Now(),
		Type:      "clear",
		Value:     12,
	})

	if len(m.Mem.VMS.Triggers) != 1 {
		t.Fatal("num VMS triggers should be 1")
	}

	m.addVMSTrigger(&report.TriggerReport[uint]{
		Timestamp: time.Now(),
		Type:      "clear",
		Value:     12,
	})

	if len(m.Mem.VMS.Triggers) != 2 {
		t.Fatal("num VMS triggers should be 2")
	}
}

func TestAddMetricErrToModule(t *testing.T) {
	m := newModule()
	m.addMetricErr("metric", errors.New("some error"))

	if len(m.Metric.Errs["metric"]) != 1 {
		t.Fatal("should have been 1 error")
	}

	m.addMetricErr("metric", errors.New("some error"))

	if len(m.Metric.Errs["metric"]) != 2 {
		t.Fatal("should have been 2 error2")
	}
}

func TestAddMetricTriggerToModule(t *testing.T) {
	m := newModule()
	m.addMetricTrigger("metric", &report.TriggerReport[float64]{
		Timestamp: time.Now(),
		Type:      "raise",
		Value:     12.12,
	})

	if len(m.Metric.Triggers["metric"]) != 1 {
		t.Fatal("should have been 1 trigger")
	}

	m.addMetricTrigger("metric", &report.TriggerReport[float64]{
		Timestamp: time.Now(),
		Type:      "raise",
		Value:     12.12,
	})

	if len(m.Metric.Triggers["metric"]) != 2 {
		t.Fatal("should have been 2 triggers")
	}
}
