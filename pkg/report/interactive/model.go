package interactive

import (
	"fmt"
	"os"
	"strings"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Message types exchanged with the bubbletea program.
type (
	opMsg struct {
		mod    string
		op     string
		ok     bool
		errStr string
	}
	logLineMsg struct{ line string }
	doneMsg    struct{}
	tickMsg    struct{ t time.Time }
)

// opStats holds the running totals for a single operation.
type opStats struct {
	executions uint
	ok         uint
	nok        uint
}

// errorEntry is a single error record shown in the errors panel.
type errorEntry struct {
	mod    string
	op     string
	errStr string
}

const (
	maxErrors   = 20
	maxLogLines = 50
	tickInterval = 2500 * time.Millisecond
)

// Styles.
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205")).
			Padding(0, 1)

	sectionTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("39"))

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")).
			Padding(0, 1)

	okStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	nokStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))

	statusRunningStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("214")).
				Bold(true)

	statusDoneStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("42")).
			Bold(true)

	logLineStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("243"))
	errorLineStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
)

// model is the bubbletea model for the interactive TUI.
type model struct {
	// stats holds per-module per-op counters.
	stats map[string]map[string]*opStats
	// errors is a capped list of encountered errors.
	errors []errorEntry
	// logLines is a capped ring of log output lines.
	logLines []string
	// done is true once the test has finished.
	done bool
	// width is the current terminal width.
	width int
}

func newModel() model {
	return model{
		stats:    make(map[string]map[string]*opStats),
		errors:   make([]errorEntry, 0, maxErrors),
		logLines: make([]string, 0, maxLogLines),
	}
}

func tickCmd() tea.Cmd {
	return tea.Tick(tickInterval, func(t time.Time) tea.Msg {
		return tickMsg{t: t}
	})
}

// Init implements tea.Model.
func (m model) Init() tea.Cmd {
	return tickCmd()
}

// Update implements tea.Model.
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			if m.done {
				// Test already finished: exit the TUI.
				return m, tea.Quit
			}
			// Test still running: re-raise SIGINT so the arbiter stops traffic.
			proc, _ := os.FindProcess(os.Getpid())
			_ = proc.Signal(syscall.SIGINT)
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width

	case tickMsg:
		return m, tickCmd()

	case opMsg:
		m.handleOp(msg)

	case logLineMsg:
		m.addLogLine(msg.line)

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
		m.addError(errorEntry{mod: msg.mod, op: msg.op, errStr: msg.errStr})
	}
}

func (m *model) addError(e errorEntry) {
	if len(m.errors) >= maxErrors {
		m.errors = m.errors[1:]
	}
	m.errors = append(m.errors, e)
}

func (m *model) addLogLine(line string) {
	if len(m.logLines) >= maxLogLines {
		m.logLines = m.logLines[1:]
	}
	m.logLines = append(m.logLines, line)
}

// View implements tea.Model.
func (m model) View() string {
	var sb strings.Builder

	// Title.
	sb.WriteString(titleStyle.Render("ARBITER — Interactive Mode"))
	sb.WriteString("\n\n")

	// Operation counts.
	sb.WriteString(sectionTitleStyle.Render("── OPERATION COUNTS ────────────────────────────────"))
	sb.WriteString("\n")
	if len(m.stats) == 0 {
		sb.WriteString("  (no operations yet)\n")
	} else {
		for modName, ops := range m.stats {
			sb.WriteString(fmt.Sprintf("  MODULE: %s\n", modName))
			var rows []string
			for opName, s := range ops {
				okStr := okStyle.Render(fmt.Sprintf("%d", s.ok))
				nokStr := nokStyle.Render(fmt.Sprintf("%d", s.nok))
				rows = append(rows,
					fmt.Sprintf("  %-24s  Exec: %6d   OK: %s   NOK: %s",
						opName, s.executions, okStr, nokStr))
			}
			sb.WriteString(boxStyle.Render(strings.Join(rows, "\n")))
			sb.WriteString("\n")
		}
	}
	sb.WriteString("\n")

	// Errors.
	sb.WriteString(sectionTitleStyle.Render(fmt.Sprintf("── ERRORS (last %d) ──────────────────────────────────", maxErrors)))
	sb.WriteString("\n")
	if len(m.errors) == 0 {
		sb.WriteString("  (none)\n")
	} else {
		for _, e := range m.errors {
			sb.WriteString(errorLineStyle.Render(fmt.Sprintf("  [%s/%s] %s", e.mod, e.op, e.errStr)))
			sb.WriteString("\n")
		}
	}
	sb.WriteString("\n")

	// Log.
	sb.WriteString(sectionTitleStyle.Render(fmt.Sprintf("── LOG (last %d lines) ───────────────────────────────", maxLogLines)))
	sb.WriteString("\n")
	if len(m.logLines) == 0 {
		sb.WriteString("  (no output yet)\n")
	} else {
		for _, line := range m.logLines {
			sb.WriteString(logLineStyle.Render("  " + line))
			sb.WriteString("\n")
		}
	}
	sb.WriteString("\n")

	// Status bar.
	if m.done {
		sb.WriteString(statusDoneStyle.Render("  ✓  Test complete! Press CTRL-C to exit."))
	} else {
		sb.WriteString(statusRunningStyle.Render("  ●  Running…  (CTRL-C to stop early)"))
	}
	sb.WriteString("\n")

	return sb.String()
}
