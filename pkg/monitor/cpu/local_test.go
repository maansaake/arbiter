package cpu

import (
	"os"
	"testing"
	"time"

	"github.com/shirou/gopsutil/v4/process"
)

func TestTempCPU(t *testing.T) {
	proc, err := process.NewProcess(int32(os.Getpid()))
	if err != nil {
		panic(err)
	}

	// Initial ignored call to avoid 0 return for first reading.
	perc, _ := proc.CPUPercent()
	t.Log(perc)
}

func TestCPU(t *testing.T) {
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
