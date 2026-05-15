// Package interactive provides a bubbletea-powered live TUI reporter that
// displays a progress bar, countdown timer, and per-module operation statistics.
package interactive

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
func New(metadata module.Metadata, totalDuration time.Duration) *Reporter {
	prog := tea.NewProgram(newModel(metadata, totalDuration), tea.WithAltScreen())
	return &Reporter{program: prog}
}

// Start implements report.Reporter. It launches the bubbletea program in a
// goroutine and sends a doneMsg when ctx is cancelled (test finished normally).
func (r *Reporter) Start(ctx context.Context) {
	go func() {
		_, _ = r.program.Run()
	}()

	go func() {
		<-ctx.Done()
		r.program.Send(doneMsg{})
	}()
}

// Op implements report.Reporter.
func (r *Reporter) Op(mod, op string, _ *module.Result, err error) {
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
