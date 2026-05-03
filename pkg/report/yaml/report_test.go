package yaml

import (
	"errors"
	"testing"
	"time"

	"github.com/maansaake/arbiter/pkg/module"
)

func TestAddOpToModule(t *testing.T) {
	m := newModuleReport()
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
