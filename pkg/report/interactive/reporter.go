// Package interactivereport provides a bubbletea-powered live TUI reporter that
// displays a progress bar, countdown timer, and per-module operation statistics.
package interactivereport

import (
	"context"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/maansaake/arbiter/pkg/module"
	"github.com/maansaake/arbiter/pkg/report"
)

// Reporter implements report.Reporter and drives a bubbletea TUI program.
type Reporter struct {
	program *tea.Program
}

var _ report.Reporter = &Reporter{}

// New creates a new Reporter initialised with module metadata and the total
// test duration so the TUI can display accurate progress and operation
// information from the start. Call Start to begin rendering.
// stopFn is called when the user presses Ctrl-C inside the TUI, allowing the
// caller to cancel the test context without relying on OS signal delivery
// (bubbletea runs the terminal in raw mode and intercepts the key event before
// the OS can raise SIGINT).
func New(metadata module.Metadata, totalDuration time.Duration, stopFn func()) *Reporter {
	return &Reporter{program: tea.NewProgram(newModel(metadata, totalDuration, stopFn), tea.WithAltScreen())}
}

// Start implements report.Reporter. It launches the bubbletea program in a
// goroutine and sends a doneMsg when ctx is cancelled (test finished normally).
func (r *Reporter) Start(ctx context.Context) {
	// bubbletea TUI go-routine, blocks until program exit via tea.Quit.
	go func() {
		_, _ = r.program.Run()
	}()

	// Reporter context monitor, once the context is cancelled the test is shutting down.
	go func() {
		<-ctx.Done()
		r.program.Send(doneMsg{})
	}()
}

// ReportOp implements report.Reporter.
func (r *Reporter) ReportOp(mod, op string, _ *module.Result, err error) {
	r.program.Send(opMsg{
		mod: mod,
		op:  op,
		ok:  err == nil,
	})
}

// Finalise implements report.Reporter. For a normally completed test it shows
// the completion footer and blocks until the user presses CTRL-C; for an
// early exit the TUI has already quit so this returns immediately.
func (r *Reporter) Finalise() error {
	r.program.Send(doneMsg{})
	r.program.Wait()
	return nil
}
