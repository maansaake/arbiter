package cpu

import (
	"os"
	"testing"
)

func TestCPU(t *testing.T) {
	mon := NewLocalCPUMonitor(int32(os.Getpid()))

	v, err := mon.Read()
	if err != nil {
		t.Fatal(err)
	}

	t.Log("read CPU percentage:", v)
	if v <= 0 {
		t.Fatal("CPU percentage less than or equal to zero, that's weird")
	}

	v, err = mon.Read()
	if err != nil {
		t.Fatal(err)
	}

	t.Log("read CPU percentage:", v)
	if v <= 0 {
		t.Fatal("CPU percentage less than or equal to zero, that's weird")
	}
}