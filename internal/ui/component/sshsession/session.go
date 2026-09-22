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
const headerHeight = 2
const messageDisplayTime = 2 * time.Second

type clearStatusMsg struct{}

type Model struct {
	term              termview.Model
	command           string
	stdErr            io.Writer
	styles            styles
	header            string
	defaultStatusText string
	currentStatusText string
}

func New(initialWidth, initialHeight int, commandAndArgs ...string) (Model, error) {
	stdErr := &utils.ProcessBufferWriter{}
	term, err := termview.New(
		termview.WithCommand(commandAndArgs[0], commandAndArgs[1:]...),
		termview.WithStdErr(stdErr),
		termview.WithInitialWidth(initialWidth),
		termview.WithInitialHeight(initialHeight-headerHeight),
	)
	if err != nil {
		return Model{}, err
	}

	statusText := strings.Join(commandAndArgs, " ")

	m := Model{
		term:              term.Focus(),
		command:           commandAndArgs[0],
		stdErr:            stdErr,
		styles:            defaultStyles(),
		defaultStatusText: statusText,
		currentStatusText: statusText,
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
		msg.Height = msg.Height - lipgloss.Height(m.headerView())
		updated, cmd := m.term.Update(msg)
		m.term = updated
		return m, cmd
	case tea.MouseMsg:
		return m.handleMouseMsg(msg)
	case termview.TextSelectedMsg:
		return m.handleTextSelectedMsg(msg)
	case clearStatusMsg:
		m.currentStatusText = m.defaultStatusText
		return m, nil
	default:
		updated, cmd := m.term.Update(msg)
		m.term = updated
		return m, cmd
	}
}

func (m Model) handleMouseMsg(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	// The alternative way would be to add another option to TermView called WithMouseYOffset(n)
	// Note: sometimes mouse events stopped working normally if console in a broken state.
	// Closing and reopening the terminal fixes it.

	h := lipgloss.Height(m.headerView())
	var adjusted tea.MouseMsg
	switch msg := msg.(type) {
	case tea.MouseClickMsg:
		msg.Y -= h
		adjusted = msg
	case tea.MouseMotionMsg:
		msg.Y -= h
		adjusted = msg
	case tea.MouseReleaseMsg:
		msg.Y -= h
		adjusted = msg
	case tea.MouseWheelMsg:
		msg.Y -= h
		adjusted = msg
	}

	if adjusted.Mouse().Y < 0 {
		// Skip mouse events in the header area.
		return m, nil
	}

	updated, cmd := m.term.Update(adjusted)
	m.term = updated
	return m, cmd
}

func (m *Model) handleTextSelectedMsg(msg termview.TextSelectedMsg) (tea.Model, tea.Cmd) {
	if msg.ID != m.term.ID() {
		return m, nil
	}

	m.currentStatusText = "Text copied to clipboard"
	return m, tea.Batch(
		tea.SetClipboard(msg.Text),
		tea.Tick(messageDisplayTime, func(time.Time) tea.Msg {
			return clearStatusMsg{}
		}),
	)
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

func (m Model) View() tea.View {
	termView := m.term.View()
	headerView := m.headerView()
	joinedView := lipgloss.JoinVertical(lipgloss.Top, headerView, termView)

	cursor := m.term.Cursor()
	if cursor != nil {
		height := lipgloss.Height(headerView)
		cursor.Y += height // Adjust cursor position for the header line
	}

	v := tea.NewView(joinedView)
	v.Cursor = cursor
	return v
}

func (m Model) headerView() string {
	statusText := m.currentStatusText
	gapSize := m.term.Width() - lipgloss.Width(m.header) - lipgloss.Width(statusText)
	gapSize = max(gapSize, 1)
	header := lipgloss.JoinHorizontal(lipgloss.Top, m.header, strings.Repeat(" ", gapSize), statusText)
	return m.styles.header.Render(header)
}

func (m *Model) SetHeader(header string) {
	m.header = header
}
