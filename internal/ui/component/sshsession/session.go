package sshsession

import (
	"io"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/grafviktor/termview"

	"github.com/grafviktor/goto/internal/constant"
	"github.com/grafviktor/goto/internal/ui/message"
	"github.com/grafviktor/goto/internal/utils"
)

type Model struct {
	term    termview.Model
	command string
	stdErr  io.Writer
}

// newTermView opens a PTY-backed terminal. Overridden in tests to avoid ConPTY
// on Windows CI (real termview.New can hang there indefinitely).
var newTermView = termview.New

func New(initialWidth, initialHeight int, commandAndArgs ...string) (Model, error) {
	stdErr := &utils.ProcessBufferWriter{}
	term, err := newTermView(
		termview.WithCommand(commandAndArgs[0], commandAndArgs[1:]...),
		termview.WithStdErr(stdErr),
		termview.WithInitialWidth(initialWidth),
		termview.WithInitialHeight(initialHeight),
	)
	if err != nil {
		return Model{}, err
	}

	m := Model{
		term:    term.Focus(),
		command: commandAndArgs[0],
		stdErr:  stdErr,
	}

	return m, nil
}

func (m Model) Init() tea.Cmd {
	return m.term.Init()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case termview.ClosedMsg:
		cmd := m.handleSessionClose(msg)
		return m, cmd
	default:
		updated, cmd := m.term.Update(msg)
		m.term = updated
		return m, cmd
	}
}

func (m Model) View() tea.View {
	v := tea.NewView(m.term.View())
	v.Cursor = m.term.Cursor()
	return v
}

func (m Model) handleSessionClose(msg termview.ClosedMsg) tea.Cmd {
	if msg.ProcessError == nil && msg.ProcessExitCode == 0 {
		return message.TeaCmd(message.RunProcessSuccess{
			ProcessType: constant.ProcessTypeSSHConnect,
			StdOut:      "",
			StdErr:      "",
		})
	}

	var errorMessage string
	// First, check if there is any output in the standard error buffer.
	if readableErrOutput, ok := m.stdErr.(*utils.ProcessBufferWriter); ok {
		errorMessage = strings.TrimSpace(string(readableErrOutput.Output))
	}

	// If there was no error message from the process, check if we can extract it from msg.
	if utils.StringEmpty(&errorMessage) {
		errorMessage = msg.ProcessError.Error()
	}

	return message.TeaCmd(message.RunProcessErrorOccurred{
		ProcessType: constant.ProcessTypeSSHConnect,
		StdOut:      "",
		StdErr:      errorMessage,
	})
}
