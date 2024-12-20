package log

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
	"tres-bon.se/arbiter/pkg/zerologr"
)

type logFile struct {
	// Full file path
	filePath string
	// Full file path, excluding the file name
	fileDir string
	// File name only
	fileName      string
	offset        int64
	lastKnownSize int64
	watcher       *fsnotify.Watcher
	handler       LogHandler
}

func NewLogFileMonitor(filePath string) Log {
	return &logFile{
		filePath:      filePath,
		lastKnownSize: -1,
	}
}

func (l *logFile) Stream(ctx context.Context, handler LogHandler) error {
	// Verify the file exists
	f, err := os.Open(l.filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	// Create a watcher on the directory of the file
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	l.fileDir = filepath.Dir(l.filePath)
	if err := watcher.Add(l.fileDir); err != nil {
		return err
	}
	l.fileName = filepath.Base(l.filePath)
	l.watcher = watcher
	l.handler = handler

	go l.monitorEvents(ctx)

	return nil
}

func (l *logFile) monitorEvents(ctx context.Context) {
	zerologr.Info("starting file monitor for new events")

	// Buffer will contain bytes from incomplete written log lines, this is
	// intentional. Once a new write event is received, the stored chunk should
	// be completed by reading more data starting from the stored offset.
	buf := &bytes.Buffer{}
	for {
		select {
		case event := <-l.watcher.Events:
			// Verify only the monitored file has changed
			if strings.Contains(event.Name, l.fileName) && event.Has(fsnotify.Write) {
				// Open the file
				f, err := os.OpenFile(l.filePath, os.O_RDONLY, 0)
				if err != nil {
					l.handler("", err)
					continue
				}
				defer f.Close()

				finfo, err := f.Stat()
				if err != nil {
					l.handler("", err)
					continue
				}

				// Continue on size match, no new content to read.
				if finfo.Size() == l.lastKnownSize {
					continue
				}

				// Looks like a log rotation has taken place, reset the offset.
				if finfo.Size() < l.lastKnownSize {
					l.offset = 0
				}
				l.lastKnownSize = finfo.Size()

				// Read from the log file and look for line breaks to parse and emit
				// everything before it.
				//
				// - Read a chunk of data
				// - Combine the chunk with rest from previous iteration, for each
				//   line break in the combo, emit the log event between start and
				//   the line break
				// - For remaining data, store in buffer for next iteration
				//
				chunk := make([]byte, 1024)
				for {
					n, readErr := f.ReadAt(chunk, l.offset)
					l.offset += int64(n)

					if readErr != nil && !errors.Is(readErr, io.EOF) {
						l.handler("", err)
						break
					}

					// Write all to buf
					buf.Write(chunk[0:n])
					for {
						// Parse for line breaks and emit log events
						line, err := buf.ReadBytes(byte('\n'))
						// No delimeter found, write the data back for later use and break.
						// Data has to be written back to the buffer since the offset has
						// already been moved up by 'n'.
						if err != nil {
							buf.Write(line)
							break
						}

						// Emit log line, minus line break
						l.handler(string(line[:len(line)-1]), nil)
					}

					if errors.Is(readErr, io.EOF) {
						break
					}
				}
			}
		case <-ctx.Done():
			zerologr.Info("closing watcher")
			l.watcher.Close()
			l.handler("", ErrStopped)
			return
		}
	}
}
