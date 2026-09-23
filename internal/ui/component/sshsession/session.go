package sshsession

import (
	"io"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/grafviktor/termview"

	"github.com/grafviktor/goto/internal/constant"
	"github.com/grafviktor/goto/internal/ui/message"
	"github.com/grafviktor/goto/internal/utils"
)

// Only used in `New`. Need to find a way to calculate this dynamically, as it may change.
const (
	statusHeight                = 2
	messageDisplayTime          = 2 * time.Second
	defaultNotificationAreaText = "ssh"
)

type clearStatusMsg struct{}

type Model struct {
	term                        termview.Model
	command                     string
	stdErr                      io.Writer
	styles                      styles
	statusLineText              string
	defaultNotificationAreaText string
	currentNotificationAreaText string
}

func New(initialWidth, initialHeight int, commandAndArgs ...string) (Model, error) {
	stdErr := &utils.ProcessBufferWriter{}
	term, err := termview.New(
		termview.WithCommand(commandAndArgs[0], commandAndArgs[1:]...),
		termview.WithStdErr(stdErr),
		termview.WithInitialWidth(initialWidth),
		termview.WithInitialHeight(initialHeight-statusHeight),
	)
	if err != nil {
		return Model{}, err
	}

	m := Model{
		term:                        term.Focus(),
		command:                     commandAndArgs[0],
		stdErr:                      stdErr,
		styles:                      defaultStyles(),
		defaultNotificationAreaText: defaultNotificationAreaText,
		currentNotificationAreaText: defaultNotificationAreaText,
	}

	return m, nil
}

func (m *Model) Init() tea.Cmd {
	return m.term.Init()
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case termview.ClosedMsg:
		cmd := m.handleSessionClose(msg)
		return m, cmd
	case tea.WindowSizeMsg:
		msg.Height -= lipgloss.Height(m.statusView())
		updated, cmd := m.term.Update(msg)
		m.term = updated
		return m, cmd
	case tea.MouseMsg:
		return m.handleMouseMsg(msg)
	case termview.TextSelectedMsg:
		return m.handleTextSelectedMsg(msg)
	case clearStatusMsg:
		m.currentNotificationAreaText = m.defaultNotificationAreaText
		return m, nil
	default:
		updated, cmd := m.term.Update(msg)
		m.term = updated
		return m, cmd
	}
}

func (m *Model) handleMouseMsg(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	if msg.Mouse().Y > m.term.Height() {
		// Skip mouse events in the status area.
		return m, nil
	}

	updated, cmd := m.term.Update(msg)
	m.term = updated
	return m, cmd
}

func (m *Model) handleTextSelectedMsg(msg termview.TextSelectedMsg) (tea.Model, tea.Cmd) {
	if msg.ID != m.term.ID() {
		return m, nil
	}

	m.currentNotificationAreaText = "Text copied to clipboard"
	return m, tea.Batch(
		tea.SetClipboard(msg.Text),
		tea.Tick(messageDisplayTime, func(time.Time) tea.Msg {
			return clearStatusMsg{}
		}),
	)
}

func (m *Model) handleSessionClose(msg termview.ClosedMsg) tea.Cmd {
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

func (m *Model) View() tea.View {
	termView := m.term.View()
	statusView := m.statusView()
	joinedView := lipgloss.JoinVertical(lipgloss.Top, termView, statusView)
	v := tea.NewView(joinedView)
	v.Cursor = m.term.Cursor()
	return v
}

func (m *Model) statusView() string {
	gapSize := m.term.Width() - lipgloss.Width(m.statusLineText) - lipgloss.Width(m.currentNotificationAreaText)
	gapSize = max(gapSize, 1)
	status := lipgloss.JoinHorizontal(lipgloss.Top, m.statusLineText, strings.Repeat(" ", gapSize),
		m.currentNotificationAreaText)
	return m.styles.statusLine.Render(status)
}

func (m *Model) SetStatus(status string) {
	m.statusLineText = status
}
