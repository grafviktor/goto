package sshsession

import (
	"io"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/grafviktor/termview"

	"github.com/grafviktor/goto/internal/constant"
	"github.com/grafviktor/goto/internal/ui/message"
	"github.com/grafviktor/goto/internal/utils"
)

type Model struct {
	term    termview.Model
	command string
	stdErr  io.Writer
	styles  styles
}

func New(initialWidth, initialHeight int, commandAndArgs ...string) (Model, error) {
	stdErr := &utils.ProcessBufferWriter{}
	term, err := termview.New(
		termview.WithCommand(commandAndArgs[0], commandAndArgs[1:]...),
		termview.WithStdErr(stdErr),
		termview.WithInitialWidth(initialWidth),
		termview.WithInitialHeight(initialHeight-2),
	)
	if err != nil {
		return Model{}, err
	}

	m := Model{
		term:    term.Focus(),
		command: commandAndArgs[0],
		stdErr:  stdErr,
		styles:  defaultStyles(),
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
	case tea.WindowSizeMsg:
		msg.Height = msg.Height - 2
		updated, cmd := m.term.Update(msg)
		m.term = updated
		return m, cmd
	default:
		updated, cmd := m.term.Update(msg)
		m.term = updated
		return m, cmd
	}
}

// func (m Model) View() tea.View {
// 	v := tea.NewView(m.term.View())
// 	v.Cursor = m.term.Cursor()
// 	return v
// }

func (m Model) View() tea.View {
	// func (m *ListModel) prefixWithGroupName(title string) string {
	// 	if !utils.StringEmpty(&m.appState.Group) {
	// 		shortGroupName := utils.StringAbbreviation(m.appState.Group)
	// 		title = m.Styles.Title.Render(title)
	// 		m.Styles.Title = m.Styles.Title.Padding(0)
	// 		return fmt.Sprintf("%s%s", m.styles.groupAbbreviation.Render(shortGroupName), title)
	// 	}

	// 	return title
	// }

	termView := m.term.View()
	statusLine := "Group: test • Host: localhost"
	joinedView := lipgloss.JoinVertical(lipgloss.Top, termView+"\n", m.styles.hostColor.Render(statusLine))

	v := tea.NewView(joinedView)
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
