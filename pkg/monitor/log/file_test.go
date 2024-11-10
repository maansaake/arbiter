package log

import (
	"context"
	"fmt"
	"os"
	"sync"
	"testing"

	"github.com/go-logr/logr"
	log "tres-bon.se/arbiter/pkg/zerologr"
)

func TestFileLogStreamStart(t *testing.T) {
	f := "./test_file.log"

	log := NewLogFileMonitor(f)
	err := log.Stream(context.TODO(), func(l string, err error) {})
	if err != nil {
		t.Fatalf("failed to start file stream: %s", err)
	}
}

func TestFileLogErrorFileDoesNotExist(t *testing.T) {
	f := "./test_file.logggg"

	log := NewLogFileMonitor(f)
	err := log.Stream(context.TODO(), func(l string, err error) {})
	if err == nil {
		t.Fatalf("file did not exist, should not have started")
	}
}

func TestFileLogEmitEvent(t *testing.T) {
	log.SetLogger(log.New(&log.Opts{Console: true}))

	fileName := "./dynamic_log_file.log"
	f, err := os.Create(fileName)
	if err != nil {
		t.Fatalf("failed to create test file: %s", err)
	}
	defer func() {
		// not much to do about the error at this stage :)
		_ = os.Remove(fileName)
		log.SetLogger(logr.Logger{})
	}()
	f.Close()

	log := NewLogFileMonitor(fileName)

	events := 100
	gotEvent := sync.WaitGroup{}
	gotEvent.Add(events)
	err = log.Stream(context.TODO(), func(logEvent string, err error) {
		if err != nil {
			t.Fatalf("unexpected log event error: %s", err)
		}

		t.Logf("log event: %s", logEvent)

		if len(logEvent) != 10 && len(logEvent) != 11 {
			t.Errorf("expected log event to be either 10 or 11 characters long, actual: %d", len(logEvent))
		}

		gotEvent.Done()
	})
	if err != nil {
		t.Fatalf("failed to start file stream: %s", err)
	}

	go func() {
		f, err := os.OpenFile(fileName, os.O_APPEND|os.O_WRONLY, 0)
		if err != nil {
			t.Errorf("failed to open file for writing: %s", err)
			return
		}
		for i := 0; i < events; i++ {
			_, err := f.Write([]byte(fmt.Sprintf("logline: %d\n", i)))
			if err != nil {
				t.Errorf("failed to write to file: %s", err)
			}
		}
	}()

	gotEvent.Wait()
}

func TestFileLogRotation(t *testing.T) {
	filePath := "./rotating.log"
	defer func() {
		// not much to do about the error at this stage :)
		_ = os.Remove(filePath)
	}()
	f, err := os.Create(filePath)
	if err != nil {
		t.Fatal("could not create rotating log file")
	}
	defer f.Close()

	log := NewLogFileMonitor(filePath)
	logS := log.(*logFile)

	events := sync.WaitGroup{}
	err = log.Stream(context.TODO(), func(s string, err error) {
		t.Log("got log event:", s)
		events.Done()
	})
	if err != nil {
		t.Fatal("failed to start log stream")
	}

	if logS.offset != 0 {
		t.Fatal("offset should have been 0")
	}

	if logS.lastKnownSize != -1 {
		t.Fatal("last known size should have been 0, was", logS.lastKnownSize)
	}

	events.Add(1)
	_, err = f.Write([]byte("a log line\n"))
	if err != nil {
		t.Fatal("failed to write to file:", err)
	}
	events.Wait()

	t.Log("last known size", logS.lastKnownSize)
	if logS.lastKnownSize < 0 {
		t.Fatal("last known size should NOT have that small")
	}

	t.Log("offset before rotate", logS.offset)
	if logS.offset < 11 {
		t.Fatal("offset should have been 11")
	}

	events.Add(1)
	// Overwrite
	err = os.WriteFile(filePath, []byte("rotated\n"), 0)
	if err != nil {
		t.Fatal("failed to write to file:", err)
	}
	events.Wait()

	t.Log("last known size", logS.lastKnownSize)
	if logS.lastKnownSize != 8 {
		t.Fatal("last known size should have been 8")
	}

	t.Log("offset after rotate", logS.offset)
	if logS.offset != 8 {
		t.Fatal("offset should have been 8")
	}
}

func TestFileLogPartialWrite(t *testing.T) {}

func TestFileLogContextDone(t *testing.T) {}
