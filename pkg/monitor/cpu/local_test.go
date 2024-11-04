package cpu

import (
	"os"
	"testing"
	"time"
)

func TestCPU(t *testing.T) {
	t.Log("CPU measurements on GitHub action runners never work")
	t.Skip()
	mon := NewLocalCPUMonitor(int32(os.Getpid()))

	v, err := mon.Read()
	if err != nil {
		t.Fatal(err)
	}

	t.Log("read CPU percentage:", v)
	if v <= 0 {
		t.Log("try again after a short wait")
		time.Sleep(100 * time.Millisecond)

		v, err = mon.Read()
		if err != nil {
			t.Fatal(err)
		}

		t.Log("read CPU percentage:", v)
		if v <= 0 {
			t.Fatal("CPU percentage less than or equal to zero, that's weird")
		}
	}
}
