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

		startTime      time.Time
		trafficEndTime time.Time // set when trafficDoneMsg is received; freezes RPM calculation
		totalDuration  time.Duration
		width          int
		height         int

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
	barWidth        = 30
	callsLabelW     = 8 // len("success:")
	timingLabelW    = 4 // len("avg:")
	headerMinGap    = 2
)

// Styles used throughout the TUI.
var (
	titleBlockStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("205")).
			PaddingLeft(3).
			PaddingRight(3).
			Bold(true).
			Foreground(lipgloss.Color("205"))

	separatorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))

	barFilledStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("42"))

	barEmptyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("238"))

	timeRemainingStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("238"))

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
			Foreground(lipgloss.Color("240"))

	opNameStyle = lipgloss.NewStyle().
			Bold(true)

	colHeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("39"))

	rateConfigStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("255"))

	modBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("39")).
			PaddingLeft(1).
			PaddingRight(1)
)

// statusBoxStyle returns a RoundedBorder box style whose border is coloured
// according to the current test state. Content styling is applied separately
// so the label and value can use different weights.
func statusBoxStyle(color lipgloss.Color) lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(color).
		PaddingLeft(1).
		PaddingRight(1)
}

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
		m.height = msg.Height

	case tickMsg:
		return m, tickCmd()

	case errMsg:
		m.errMsg = msg.err.Error()

	case opMsg:
		m.handleOp(msg)

	case trafficDoneMsg:
		m.trafficDone = true
		m.trafficEndTime = time.Now()

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
		modInnerW := w - 4 // border(2) + padding(2) consumed by modBoxStyle
		sb.WriteString(m.renderModule(mod, modInnerW))
		sb.WriteString("\n")
	}

	footer := m.renderFooter()
	if footer != "" {
		if m.height > 0 {
			contentLines := strings.Count(sb.String(), "\n") + 1
			footerLines := strings.Count(footer, "\n") + 1
			if padding := m.height - contentLines - footerLines; padding > 0 {
				sb.WriteString(strings.Repeat("\n", padding))
			}
		}
		sb.WriteString(footer)
		sb.WriteString("\n")
	}

	return sb.String()
}

// renderFooter returns the status message shown at the bottom of the screen.
func (m *model) renderFooter() string {
	switch {
	case m.errMsg != "":
		return doneStyle.Render("Error: " + m.errMsg)
	case m.done:
		return doneStyle.Render("Test complete! Press CTRL-C to exit.")
	case m.trafficDone:
		return doneStyle.Render("Ramping down...")
	default:
		return ""
	}
}

// renderHeader renders the top bar: title block on the left, a state-coloured
// status box next, then the progress bar with countdown stacked below it
// pushed to roughly the middle of the remaining space. The bar width is at
// least 50% of the space available after the title and status boxes.
func (m *model) renderHeader(w int) string {
	titleBlock := titleBlockStyle.Render("arbiter")

	var statusColor lipgloss.Color
	var statusText string
	switch {
	case m.errMsg != "":
		statusColor = lipgloss.Color("196")
		statusText = "ERROR"
	case m.done:
		statusColor = lipgloss.Color("214")
		statusText = "DONE"
	default:
		statusColor = lipgloss.Color("42")
		statusText = "RUNNING"
	}
	statusLabel := lipgloss.NewStyle().Foreground(statusColor).Render("test status")
	statusValue := lipgloss.NewStyle().Bold(true).Foreground(statusColor).Render(statusText)
	statusBox := statusBoxStyle(statusColor).Render(statusLabel + "  " + statusValue)

	titleW := lipgloss.Width(titleBlock)
	statusW := lipgloss.Width(statusBox)

	// Bar width: at least 50% of the space remaining after title + status boxes.
	rightAvail := w - titleW - headerMinGap - statusW
	barW := rightAvail / 2
	if barW < barWidth {
		barW = barWidth
	}

	bar := m.renderProgressBar(barW)
	timeRaw := timeRemainingStyle.Render(formatDuration(m.timeRemaining()) + " remaining")
	centeredTime := lipgloss.NewStyle().Width(barW).Align(lipgloss.Center).Render(timeRaw)
	rightSection := bar + "\n" + centeredTime
	rightW := lipgloss.Width(rightSection)

	// Spacer pushes the bar to the midpoint of whatever space is left.
	available := w - titleW - headerMinGap - statusW - rightW
	spacer := available / 2
	if spacer < headerMinGap {
		spacer = headerMinGap
	}

	return lipgloss.JoinHorizontal(lipgloss.Center,
		titleBlock,
		strings.Repeat(" ", headerMinGap),
		statusBox,
		strings.Repeat(" ", spacer),
		rightSection,
	)
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

// renderModule renders a full module section — header line plus all operation
// boxes — wrapped in a rounded border box. contentW is the inner content width
// (the box border and padding are added on top).
func (m *model) renderModule(mod *module.Meta, contentW int) string {
	var sb strings.Builder
	sb.WriteString(modHeaderStyle.Render("Module: " + mod.Name()))
	sb.WriteString("\n\n")

	opInnerW := contentW - 4 // border(2) + padding(2) consumed by opBoxStyle
	twoCol := false
	if halfOpInnerW := (contentW - 10) / 2; halfOpInnerW >= 30 {
		opInnerW = halfOpInnerW
		twoCol = true
	}

	ops := mod.Ops()
	for i := 0; i < len(ops); {
		box1 := m.renderOp(mod.Name(), ops[i], opInnerW)
		i++
		if twoCol && i < len(ops) {
			box2 := m.renderOp(mod.Name(), ops[i], opInnerW)
			i++
			sb.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, box1, "  ", box2))
		} else {
			sb.WriteString(box1)
		}
		sb.WriteString("\n")
	}

	return modBoxStyle.Width(contentW).Render(strings.TrimSuffix(sb.String(), "\n"))
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

	elapsed := time.Since(m.startTime)
	if m.trafficDone {
		elapsed = m.trafficEndTime.Sub(m.startTime)
	}

	if modStats, ok := m.stats[modName]; ok {
		if stats, opOK := modStats[op.Name]; opOK {
			executions = stats.executions
			nok = stats.nok
			okCount = stats.ok
			rpm = stats.observedRPM(elapsed)
			if executions > 0 && stats.totalDuration > 0 {
				//nolint:gosec // no risk of overflow since the total duration is the sum
				avgDur = stats.totalDuration / time.Duration(executions)
				minDur = stats.minDuration
				maxDur = stats.maxDuration
			}
		}
	}

	// Three side-by-side columns: Rate | Calls | Timing
	colW := (innerW - 2) / 3 // 2 single-char gaps between 3 columns
	if colW < 8 {
		colW = 8
	}
	colStyle := lipgloss.NewStyle().Width(colW)

	// Rate: configured rate in the header; "Rate" is bold-blue, the value is plain white.
	rateCol := colHeaderStyle.Render("Rate") + rateConfigStyle.Render(fmt.Sprintf(" (%d/min)", op.Rate)) + "\n" +
		fmt.Sprintf("actual: %d/min", rpm)

	// Calls: labels padded to callsLabelW so values align.
	callsCol := colHeaderStyle.Render("Calls") + "\n" +
		fmt.Sprintf("%-*s %d", callsLabelW, "calls:", executions) + "\n" +
		fmt.Sprintf("%-*s %d", callsLabelW, "failed:", nok) + "\n" +
		fmt.Sprintf("%-*s %s", callsLabelW, "success:", successStr(executions, okCount))

	// Timing: labels padded to timingLabelW so values align; colon on each label.
	timingCol := colHeaderStyle.Render("Timing") + "\n" +
		fmt.Sprintf("%-*s %s", timingLabelW, "avg:", formatOpDuration(avgDur)) + "\n" +
		fmt.Sprintf("%-*s %s", timingLabelW, "min:", formatOpDuration(minDur)) + "\n" +
		fmt.Sprintf("%-*s %s", timingLabelW, "max:", formatOpDuration(maxDur))

	columns := lipgloss.JoinHorizontal(lipgloss.Top,
		colStyle.Render(rateCol), " ",
		colStyle.Render(callsCol), " ",
		colStyle.Render(timingCol))

	return opBoxStyle.Width(innerW).Render(
		opNameStyle.Render("Operation: "+op.Name) + "\n\n" + columns,
	)
}

// successStr returns a formatted success percentage, or "—" when no calls
// have been made yet.
func successStr(executions, ok uint) string {
	if executions == 0 {
		return "—"
	}

	return fmt.Sprintf("%.1f%%", float64(ok)/float64(executions)*100)
}

// observedRPM returns the actual observed rate per minute. elapsed is the
// duration since the test started and may be frozen when traffic has stopped,
// preventing the rate from declining after the test ends.
func (s *opStats) observedRPM(elapsed time.Duration) uint {
	if s.executions == 0 {
		return 0
	}

	minutes := elapsed.Minutes()
	if minutes < 0.001 {
		return 0
	}

	return uint(math.Round(float64(s.executions) / minutes))
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
