package memory

import (
	"os"
	"testing"
)

func TestRSSVMS(t *testing.T) {
	mon := NewLocalMemoryMonitor(int32(os.Getpid()))

	v, err := mon.RSS()
	if err != nil {
		t.Fatal(err)
	}

	t.Log("read RSS:", v)
	if v <= 0 {
		t.Fatal("RSS less than or equal to zero, that's weird")
	}

	v, err = mon.VMS()
	if err != nil {
		t.Fatal(err)
	}

	t.Log("read VMS:", v)
	if v <= 0 {
		t.Fatal("VMS less than or equal to zero, that's weird")
	}
}
