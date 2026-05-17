//nolint:gochecknoglobals,mnd // model has a lot of global styles
package interactivereport

import (
	"fmt"
	"math"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/maansaake/arbiter/pkg/module"
)

type (
	// Message types exchanged with the bubbletea program.

	// opMsg is sent when an operation is executed, containing the module and operation
	// name and whether the call was successful.
	opMsg struct {
		mod      string
		op       string
		ok       bool
		duration time.Duration
	}
	// errMsg is sent when an error is reported via ReportError.
	errMsg struct {
		err error
	}
	// trafficDoneMsg is sent when the traffic context is cancelled.
	trafficDoneMsg struct{}
	// doneMsg is sent when the test completes.
	doneMsg struct{}
	// tickMsg is sent at regular intervals to trigger screen refreshes and progress bar updates.
	tickMsg struct{ t time.Time }

	// model is the bubbletea model for the interactive TUI.
	model struct {
		metadata module.Metadata
		stats    map[string]map[string]*opStats

		errMsg      string
		trafficDone bool
		done        bool

		startTime     time.Time
		totalDuration time.Duration
		width         int

		// trafficCancel is called when the user presses Ctrl-C inside the TUI so that
		// the arbiter shutdown sequence is triggered without relying on SIGINT
		// delivery (bubbletea runs in raw terminal mode and consumes the key
		// event before the OS raises the signal).
		trafficCancel func()
	}

	// opStats holds the running totals for a single operation.
	opStats struct {
		executions    uint
		ok            uint
		nok           uint
		totalDuration time.Duration
		minDuration   time.Duration
		maxDuration   time.Duration
	}
)

const (
	refreshInterval = time.Second
	defaultWidth    = 80
	opNameWidth     = 20
	opColWidth      = 20
)

// Styles used throughout the TUI.
var (
	brandStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205"))

	separatorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))

	barFilledStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("42"))

	barEmptyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("238"))

	timeRemainingStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("39"))

	modHeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("39"))

	opBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")).
			PaddingLeft(1).
			PaddingRight(1)

	opDisabledBoxStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("237")).
				PaddingLeft(1).
				PaddingRight(1)

	opDisabledTextStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("238"))

	doneStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("214"))

	opNameStyle = lipgloss.NewStyle().
			Bold(true)
)

// newModel creates a model pre-populated with module and operation metadata.
func newModel(metadata module.Metadata, d time.Duration, stopFn func()) *model {
	return &model{
		metadata:      metadata,
		stats:         make(map[string]map[string]*opStats),
		trafficCancel: stopFn,
		startTime:     time.Now(),
		totalDuration: d,
	}
}

func tickCmd() tea.Cmd {
	return tea.Tick(refreshInterval, func(t time.Time) tea.Msg {
		return tickMsg{t: t}
	})
}

// Init implements tea.Model.
func (m *model) Init() tea.Cmd {
	return tickCmd()
}

// Update implements tea.Model.
func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			// Call traffic cancel in case this is a premature interrupt.
			m.trafficCancel()
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width

	case tickMsg:
		return m, tickCmd()

	case errMsg:
		m.errMsg = msg.err.Error()

	case opMsg:
		m.handleOp(msg)

	case trafficDoneMsg:
		m.trafficDone = true

	case doneMsg:
		m.done = true
	}

	return m, nil
}

func (m *model) handleOp(msg opMsg) {
	if _, ok := m.stats[msg.mod]; !ok {
		m.stats[msg.mod] = make(map[string]*opStats)
	}

	stats, ok := m.stats[msg.mod][msg.op]
	if !ok {
		stats = &opStats{}
		m.stats[msg.mod][msg.op] = stats
	}

	stats.executions++
	if msg.ok {
		stats.ok++
	} else {
		stats.nok++
	}

	if msg.duration > 0 {
		stats.totalDuration += msg.duration
		if stats.minDuration == 0 || msg.duration < stats.minDuration {
			stats.minDuration = msg.duration
		}
		if msg.duration > stats.maxDuration {
			stats.maxDuration = msg.duration
		}
	}
}

// View implements tea.Model.
func (m *model) View() string {
	w := m.width
	if w == 0 {
		w = defaultWidth
	}

	var sb strings.Builder

	sb.WriteString(m.renderHeader(w))
	sb.WriteString("\n")
	sb.WriteString(separatorStyle.Render(strings.Repeat("─", w)))
	sb.WriteString("\n\n")

	for _, mod := range m.metadata {
		sb.WriteString(modHeaderStyle.Render(mod.Name()))
		sb.WriteString("\n")

		innerW := w - 4
		twoCol := false
		if halfInnerW := (w - 10) / 2; halfInnerW >= 30 {
			innerW = halfInnerW
			twoCol = true
		}

		ops := mod.Ops()
		for i := 0; i < len(ops); {
			box1 := m.renderOp(mod.Name(), ops[i], innerW)
			i++
			if twoCol && i < len(ops) {
				box2 := m.renderOp(mod.Name(), ops[i], innerW)
				i++
				sb.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, box1, "  ", box2))
			} else {
				sb.WriteString(box1)
			}
			sb.WriteString("\n")
		}
		sb.WriteString("\n")
	}

	// err -> done -> traffic done
	switch {
	case m.errMsg != "":
		sb.WriteString(doneStyle.Render("Error: " + m.errMsg))
		sb.WriteString("\n")
	case m.done:
		sb.WriteString(doneStyle.Render("Test complete! Press CTRL-C to exit."))
		sb.WriteString("\n")
	case m.trafficDone:
		sb.WriteString(doneStyle.Render("Ramping down..."))
		sb.WriteString("\n")
	}

	return sb.String()
}

// renderHeader renders the top bar: "arbiter" brand on the left, progress bar
// and countdown clock filling the remaining width.
func (m *model) renderHeader(w int) string {
	brand := brandStyle.Render("arbiter")
	suffix := " " + timeRemainingStyle.Render(formatDuration(m.timeRemaining())+" remaining")

	brandW := lipgloss.Width(brand)
	suffixW := lipgloss.Width(suffix)
	barW := w - brandW - suffixW - 1 // 1 space between brand and bar
	if barW < 1 {
		barW = 1
	}

	return brand + " " + m.renderProgressBar(barW) + suffix
}

// renderProgressBar renders a filled/empty Unicode block progress bar.
func (m *model) renderProgressBar(w int) string {
	filled := int(math.Round(m.progressRatio() * float64(w)))
	if filled > w {
		filled = w
	}

	empty := w - filled

	return barFilledStyle.Render(strings.Repeat("█", filled)) +
		barEmptyStyle.Render(strings.Repeat("░", empty))
}

// progressRatio returns a value in [0.0, 1.0] representing elapsed test time.
func (m *model) progressRatio() float64 {
	if m.done || m.totalDuration == 0 {
		return 1.0
	}

	p := time.Since(m.startTime).Seconds() / m.totalDuration.Seconds()
	if p > 1.0 {
		return 1.0
	}

	return p
}

// timeRemaining returns the time left until the test completes.
func (m *model) timeRemaining() time.Duration {
	if m.done {
		return 0
	}

	rem := time.Until(m.startTime.Add(m.totalDuration))
	if rem < 0 {
		return 0
	}

	return rem
}

// renderOp renders a single operation's statistics box. Disabled operations
// are rendered with a greyed-out border and [DISABLED] label.
func (m *model) renderOp(modName string, op *module.Op, innerW int) string {
	if op.Disabled {
		content := opDisabledTextStyle.Render(
			fmt.Sprintf("%-*s  [DISABLED]", opNameWidth, op.Name),
		)

		return opDisabledBoxStyle.Width(innerW).Render(content)
	}

	var (
		executions, nok, okCount, rpm uint
		avgDur, minDur, maxDur        time.Duration
	)
	if modStats, ok := m.stats[modName]; ok {
		stats := modStats[op.Name]
		executions = stats.executions
		nok = stats.nok
		okCount = stats.ok
		rpm = stats.observedRPM(m.startTime)
		if executions > 0 && stats.totalDuration > 0 {
			//nolint:gosec // no risk of overflow since the total duration is the sum
			avgDur = stats.totalDuration / time.Duration(executions)
			minDur = stats.minDuration
			maxDur = stats.maxDuration
		}
	}

	line1 := fmt.Sprintf("%-*s  %s",
		opColWidth, fmt.Sprintf("rate: %d/min", op.Rate),
		fmt.Sprintf("actual: %d/min", rpm))
	line2 := fmt.Sprintf("%-*s  %-*s  %s",
		opColWidth, fmt.Sprintf("calls: %d", executions),
		opColWidth, fmt.Sprintf("failed: %d", nok),
		fmt.Sprintf("success: %s", successStr(executions, okCount)))
	line3 := fmt.Sprintf("%-*s  %-*s  %s",
		opColWidth, fmt.Sprintf("avg %s", formatOpDuration(avgDur)),
		opColWidth, fmt.Sprintf("min %s", formatOpDuration(minDur)),
		fmt.Sprintf("max %s", formatOpDuration(maxDur)))

	return opBoxStyle.Width(innerW).Render(opNameStyle.Render(op.Name) + "\n" + line1 + "\n" + line2 + "\n" + line3)
}

// successStr returns a formatted success percentage, or "—" when no calls
// have been made yet.
func successStr(executions, ok uint) string {
	if executions == 0 {
		return "—"
	}

	return fmt.Sprintf("%.1f%%", float64(ok)/float64(executions)*100)
}

// observedRPM returns the actual observed rate per minute since the first call.
func (s *opStats) observedRPM(startTime time.Time) uint {
	if s.executions == 0 {
		return 0
	}

	elapsed := time.Since(startTime).Minutes()
	if elapsed < 0.001 {
		return 0
	}

	return uint(math.Round(float64(s.executions) / elapsed))
}

// formatOpDuration formats an operation duration in a human-readable short form.
func formatOpDuration(d time.Duration) string {
	if d == 0 {
		return "—"
	}

	if d >= time.Second {
		return fmt.Sprintf("%.2fs", d.Seconds())
	}

	if d >= time.Millisecond {
		return fmt.Sprintf("%.0fms", float64(d)/float64(time.Millisecond))
	}

	if d >= time.Microsecond {
		return fmt.Sprintf("%.0fµs", float64(d)/float64(time.Microsecond))
	}

	return fmt.Sprintf("%dns", d.Nanoseconds())
}

// formatDuration formats a duration as MM:SS or H:MM:SS.
func formatDuration(d time.Duration) string {
	if d < 0 {
		d = 0
	}

	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60

	if h > 0 {
		return fmt.Sprintf("%d:%02d:%02d", h, m, s)
	}

	return fmt.Sprintf("%02d:%02d", m, s)
}
