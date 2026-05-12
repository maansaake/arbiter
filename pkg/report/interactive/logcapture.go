package interactive

import (
	"bufio"
	"io"

	tea "github.com/charmbracelet/bubbletea"
)

// logCapture is an io.Writer that splits input into lines and sends each line
// as a logLineMsg to a bubbletea program.
type logCapture struct {
	pw      *io.PipeWriter
	pr      *io.PipeReader
	program *tea.Program
}

func newLogCapture(program *tea.Program) *logCapture {
	pr, pw := io.Pipe()
	lc := &logCapture{pr: pr, pw: pw, program: program}
	go lc.scan()
	return lc
}

// Write implements io.Writer.
func (lc *logCapture) Write(p []byte) (int, error) {
	return lc.pw.Write(p)
}

func (lc *logCapture) close() {
	_ = lc.pw.Close()
}

func (lc *logCapture) scan() {
	scanner := bufio.NewScanner(lc.pr)
	for scanner.Scan() {
		lc.program.Send(logLineMsg{line: scanner.Text()})
	}
}
