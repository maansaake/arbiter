// Package interactive provides a bubbletea-powered live TUI reporter that
// displays operation counts, errors, and log output at regular intervals.
package interactive

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-logr/logr"
	"github.com/maansaake/arbiter/pkg/module"
	"github.com/maansaake/arbiter/pkg/report"
)

// Reporter implements report.Reporter and drives a bubbletea TUI program that
// shows live operation counts, errors, and streamed log output.
type Reporter struct {
	program    *tea.Program
	logCapture *logCapture
}

var _ report.Reporter = &Reporter{}

// New creates a new interactive Reporter. The bubbletea program is not yet
// started; call Start to begin rendering.
func New() *Reporter {
	m := newModel()
	// The program must exist before newLogCapture so that Send can be called
	// from the scan goroutine that starts inside newLogCapture.
	prog := tea.NewProgram(m, tea.WithAltScreen())
	lc := newLogCapture(prog)
	return &Reporter{
		program:    prog,
		logCapture: lc,
	}
}

// Start implements report.Reporter. It launches the bubbletea program in a
// goroutine and watches ctx so that a doneMsg is sent when the test finishes.
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
	var errStr string
	if err != nil {
		errStr = err.Error()
	}

	r.program.Send(opMsg{
		mod:    mod,
		op:     op,
		ok:     err == nil,
		errStr: errStr,
	})
}

// Finalise implements report.Reporter. It marks the test as done in the TUI
// and blocks until the user presses CTRL-C to exit.
func (r *Reporter) Finalise() error {
	r.program.Send(doneMsg{})
	r.program.Wait()
	r.logCapture.close()
	return nil
}

// NewLogger returns a logr.Logger that writes formatted log lines into the
// TUI log panel. Call this after New() and set it as the global zerologr
// logger so all arbiter log output appears in the TUI.
func (r *Reporter) NewLogger() logr.Logger {
	return logr.New(&interactiveSink{capture: r.logCapture})
}
