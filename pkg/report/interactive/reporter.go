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

// reporter implements report.reporter and drives a bubbletea TUI program.
type reporter struct {
	program *tea.Program

	// trafficCtx is used to monitor the traffic progression, to display helpful
	// messages in the TUI.
	trafficCtx context.Context
}

var _ report.Reporter = &reporter{}

// New creates a new Reporter initialised with module metadata and the total
// test duration so the TUI can display accurate progress and operation
// information from the start. Call Start to begin rendering.
// stopFn is called when the user presses Ctrl-C inside the TUI, allowing the
// caller to cancel the test context without relying on OS signal delivery
// (bubbletea runs the terminal in raw mode and intercepts the key event before
// the OS can raise SIGINT).
func New(
	metadata module.Metadata,
	totalDuration time.Duration,
	//nolint:revive // the traffic context is special and not releated to the function really
	trafficCtx context.Context, trafficCancel func(),
) report.Reporter {
	return &reporter{
		program:    tea.NewProgram(newModel(metadata, totalDuration, trafficCancel), tea.WithAltScreen()),
		trafficCtx: trafficCtx,
	}
}

// Start implements report.Reporter. It launches the bubbletea program in a
// goroutine and sends a doneMsg when ctx is cancelled (test finished normally).
func (r *reporter) Start(reporterCtx context.Context) {
	// bubbletea TUI go-routine, blocks until program exit via tea.Quit.
	go func() {
		_, _ = r.program.Run()
	}()

	// Reporter context monitor, once the context is cancelled the test is shutting down.
	go func() {
		for {
			select {
			case <-reporterCtx.Done():
				r.program.Send(doneMsg{})
				return
			case <-r.trafficCtx.Done():
				r.program.Send(trafficDoneMsg{})
			}
		}
	}()
}

func (r *reporter) ReportError(err error) {
	r.program.Send(errMsg{err: err})
}

// ReportOp implements report.Reporter.
func (r *reporter) ReportOp(mod, op string, _ *module.Result, err error) {
	r.program.Send(opMsg{
		mod: mod,
		op:  op,
		ok:  err == nil,
	})
}

// Finalise implements report.Reporter. For a normally completed test it shows
// the completion footer and blocks until the user presses CTRL-C; for an
// early exit the TUI has already quit so this returns immediately.
func (r *reporter) Finalise() error {
	r.program.Wait()
	return nil
}
