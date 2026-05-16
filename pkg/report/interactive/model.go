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
		mod string
		op  string
		ok  bool
	}
	// doneMsg is sent when the test completes.
	doneMsg struct{}
	// tickMsg is sent at regular intervals to trigger screen refreshes and progress bar updates.
	tickMsg struct{ t time.Time }

	// model is the bubbletea model for the interactive TUI.
	model struct {
		metadata      module.Metadata
		stats         map[string]map[string]*opStats
		done          bool
		startTime     time.Time
		totalDuration time.Duration
		width         int
	}

	// opStats holds the running totals for a single operation.
	opStats struct {
		executions uint
		ok         uint
		nok        uint
	}
)

const (
	refreshInterval = time.Second
	defaultWidth    = 80
	opNameWidth     = 20
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
)

// newModel creates a model pre-populated with module and operation metadata.
func newModel(metadata module.Metadata, d time.Duration) *model {
	return &model{
		metadata:      metadata,
		stats:         make(map[string]map[string]*opStats),
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
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width

	case tickMsg:
		return m, tickCmd()

	case opMsg:
		m.handleOp(msg)

	case doneMsg:
		m.done = true
	}

	return m, nil
}

func (m *model) handleOp(msg opMsg) {
	if _, ok := m.stats[msg.mod]; !ok {
		m.stats[msg.mod] = make(map[string]*opStats)
	}

	s, ok := m.stats[msg.mod][msg.op]
	if !ok {
		s = &opStats{}
		m.stats[msg.mod][msg.op] = s
	}

	s.executions++
	if msg.ok {
		s.ok++
	} else {
		s.nok++
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
		for _, op := range mod.Ops() {
			sb.WriteString(m.renderOp(mod.Name(), op, w))
			sb.WriteString("\n")
		}
		sb.WriteString("\n")
	}

	if m.done {
		sb.WriteString(doneStyle.Render("Test complete! Press CTRL-C to exit."))
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
func (m *model) renderOp(modName string, op *module.Op, w int) string {
	// Inner width: total minus 2 border chars and 2 padding chars.
	innerW := w - 4

	if op.Disabled {
		content := opDisabledTextStyle.Render(
			fmt.Sprintf("%-*s  [DISABLED]", opNameWidth, op.Name),
		)

		return opDisabledBoxStyle.Width(innerW).Render(content)
	}

	var stats *opStats
	if modStats, ok := m.stats[modName]; ok {
		stats = modStats[op.Name]
	}

	var executions, nok, okCount uint
	if stats != nil {
		executions = stats.executions
		nok = stats.nok
		okCount = stats.ok
	}

	content := fmt.Sprintf(
		"%-*s  rate: %d/min   calls: %d   failed: %d   success: %s",
		opNameWidth, op.Name, op.Rate, executions, nok, successStr(executions, okCount),
	)

	return opBoxStyle.Width(innerW).Render(content)
}

// successStr returns a formatted success percentage, or "—" when no calls
// have been made yet.
func successStr(executions, ok uint) string {
	if executions == 0 {
		return "—"
	}

	return fmt.Sprintf("%.1f%%", float64(ok)/float64(executions)*100)
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
