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

	if v <= 0 {
		t.Fatal("RSS less than or equal to zero, that's weird")
	}
	t.Log("read RSS:", v)

	v, err = mon.VMS()
	if err != nil {
		t.Fatal(err)
	}

	if v <= 0 {
		t.Fatal("VMS less than or equal to zero, that's weird")
	}
	t.Log("read VMS:", v)
}
