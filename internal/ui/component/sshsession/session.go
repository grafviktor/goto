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

func New(command string, args ...string) (Model, error) {
	stdErr := &utils.ProcessBufferWriter{}
	term, err := termview.New(
		termview.WithCommand(command, args...),
		termview.WithStdErr(stdErr),
	)
	if err != nil {
		return Model{}, err
	}

	return Model{term: term.Focus(), command: command, stdErr: stdErr}, nil
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
	// return tea.NewView("Running: " + m.command)
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

	var readableStdErr string
	if readableErrOutput, ok := m.stdErr.(*utils.ProcessBufferWriter); ok {
		readableStdErr = strings.TrimSpace(string(readableErrOutput.Output))
	}

	return message.TeaCmd(message.RunProcessErrorOccurred{
		ProcessType: constant.ProcessTypeSSHConnect,
		StdOut:      "",
		StdErr:      readableStdErr,
	})
}
