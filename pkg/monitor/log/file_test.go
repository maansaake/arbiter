package log

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/go-logr/logr"
	"tres-bon.se/arbiter/pkg/zerologr"
)

func TestFileLogStreamStart(t *testing.T) {
	fileName := "./test_file.log"
	_, err := os.Create(fileName)
	if err != nil {
		t.Fatalf("failed to create test file: %s", err)
	}
	defer func() {
		// not much to do about the error at this stage :)
		_ = os.Remove(fileName)
	}()

	log := NewLogFileMonitor(fileName)
	err = log.Stream(context.TODO(), func(l string, err error) {})
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
	fileName := "./dynamic_log_file.log"
	f, err := os.Create(fileName)
	if err != nil {
		t.Fatalf("failed to create test file: %s", err)
	}
	defer func() {
		// not much to do about the error at this stage :)
		_ = os.Remove(fileName)
	}()
	f.Close()

	log := NewLogFileMonitor(fileName)

	events := 100
	gotEvent := sync.WaitGroup{}
	gotEvent.Add(events)
	ctx, cancel := context.WithCancel(context.Background())
	err = log.Stream(ctx, func(logEvent string, err error) {
		if errors.Is(err, ErrStopped) {
			return
		}
		if err != nil {
			t.Fatalf("unexpected log event error: %v", err)
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
	cancel()
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

func TestFileLogPartialWrite(t *testing.T) {
	filePath := "./partial.log"
	defer func() {
		// not much to do about the error at this stage :)
		_ = os.Remove(filePath)
	}()
	f, err := os.Create(filePath)
	if err != nil {
		t.Fatal("could not create partial log file")
	}
	defer f.Close()

	log := NewLogFileMonitor(filePath)
	logS := log.(*logFile)

	events := sync.WaitGroup{}
	err = log.Stream(context.TODO(), func(s string, err error) {
		t.Log("got log event:", s)

		if s != "a partial logline now completed" {
			t.Fatal("log message contents not as expected")
		}

		events.Done()
	})
	if err != nil {
		t.Fatal("failed to start log stream")
	}

	if logS.offset != 0 {
		t.Fatal("offset should have been 0")
	}

	// Don't await until line break is written or the test hangs.
	events.Add(1)
	_, err = f.Write([]byte("a partial log"))
	if err != nil {
		t.Fatal("failed to write to file:", err)
	}

	_, err = f.Write([]byte("line now completed\n"))
	if err != nil {
		t.Fatal("failed to write to file:", err)
	}

	// Await event emitter
	events.Wait()
}

func TestFileLogContextDone(t *testing.T) {
	filePath := "./context.log"
	zerologr.SetLogger(zerologr.New(&zerologr.Opts{Console: true}))
	defer func() {
		// not much to do about the error at this stage :)
		_ = os.Remove(filePath)
		zerologr.SetLogger(logr.Logger{})
	}()
	f, err := os.Create(filePath)
	if err != nil {
		t.Fatal("could not create context log file")
	}
	defer f.Close()

	log := NewLogFileMonitor(filePath)
	logS := log.(*logFile)

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	err = log.Stream(ctx, func(s string, err error) {})
	if err != nil {
		t.Fatal("failed to start log stream")
	}
	t.Log(len(logS.watcher.WatchList()))

	cancel()

	att := 0
	for {
		if len(logS.watcher.WatchList()) == 0 {
			break
		}
		att += 1

		if att == 50 {
			t.Fatal("watch list not cleared as expected")
			break
		}

		time.Sleep(1 * time.Millisecond)
	}
}
