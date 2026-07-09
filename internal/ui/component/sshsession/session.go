package sshsession

import (
	tea "charm.land/bubbletea/v2"
	"github.com/grafviktor/goto/internal/constant"
	"github.com/grafviktor/goto/internal/ui/message"
	"github.com/grafviktor/termview"
)

type Model struct {
	term    termview.Model
	command string
}

func New(command string, args ...string) Model {
	term, err := termview.New(termview.WithCommand(command, args...))
	if err != nil {
		// dispatch run process error message to main model
	}

	return Model{term: term.Focus(), command: command}
}

func (m Model) Init() tea.Cmd {
	return m.term.Init()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case termview.ClosedMsg:
		return m, func() tea.Msg {
			return message.RunProcessSuccess{
				ProcessType: constant.ProcessTypeSSHConnect,
				StdOut:      "exited",
				StdErr:      "",
			}
		}
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
