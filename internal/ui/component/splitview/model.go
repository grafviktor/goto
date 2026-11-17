package splitview

import (
	tea "charm.land/bubbletea/v2"
)

const (
	sessionHostListMinWidth  = 28
	sessionHostListMaxWidth  = 42
	sessionMinWidth          = 20
	sessionSplitDividerWidth = 1
)

type Model struct {
	left  tea.Model
	right tea.Model
}

func New(left, right tea.Model) *Model {
	return &Model{
		left:  left,
		right: right,
	}
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd
	// if input event {
	switch msg := msg.(type) {
	case tea.KeyMsg, tea.MouseMsg:
		m.left, cmd = m.left.Update(msg)
		cmds = append(cmds, cmd)
	case tea.WindowSizeMsg:
		leftWidth, rightWidth := sessionPaneWidths(msg.Width)
		msg.Width = leftWidth
		m.left, cmd = m.left.Update(msg)
		cmds = append(cmds, cmd)
		msg.Width = rightWidth
		m.right, cmd = m.right.Update(msg)
		cmds = append(cmds, cmd)
	default:
		m.left, cmd = m.left.Update(msg)
		cmds = append(cmds, cmd)
		m.right, cmd = m.right.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *Model) View() tea.View {
	return tea.NewView(":-)")
}

// sessionPaneWidths splits the terminal between the host list, a one-column divider, and the session.
func sessionPaneWidths(total int) (int, int) {
	var listWidth, sessionWidth int
	if total <= 0 {
		return 0, 0
	}

	usable := total - sessionSplitDividerWidth
	if usable < 2 { //nolint:mnd // a list column and a session column
		if total == 1 {
			return 1, 0
		}

		return 1, total - 1 - sessionSplitDividerWidth
	}

	listWidth = usable / 3 //nolint:mnd // the host list takes about one third of the screen
	if listWidth < sessionHostListMinWidth {
		listWidth = sessionHostListMinWidth
	}
	if listWidth > sessionHostListMaxWidth {
		listWidth = sessionHostListMaxWidth
	}

	sessionWidth = usable - listWidth
	if sessionWidth < sessionMinWidth {
		sessionWidth = sessionMinWidth
		listWidth = usable - sessionWidth
	}
	if listWidth < 1 {
		listWidth = 1
		sessionWidth = usable - listWidth
	}
	if sessionWidth < 1 {
		sessionWidth = 1
		listWidth = max(usable-sessionWidth, 1)
	}

	return listWidth, sessionWidth
}

func (m *Model) handleSessionInput(msg tea.Msg) (tea.Model, tea.Cmd) {
	// keyMsg, isKey := msg.(tea.KeyPressMsg)
	// if isKey && m.embeddedSessionUsesHostList() && keyMsg.Keystroke() == sessionHostListKey {
	// 	open := !m.sessionSplit
	// 	if open {
	// 		m.logger.Debug("[UI] Show the host list beside the SSH session")
	// 	} else {
	// 		m.logger.Debug("[UI] Hide the host list beside the SSH session")
	// 	}

	// 	return m, m.setSessionSplit(open)
	// }

	/*if m.sessionSplit {
		if mouseMsg, ok := msg.(tea.MouseMsg); ok {
			return m, m.handleSessionSplitMouse(mouseMsg)
		}

		var cmd tea.Cmd
		m.modelHostList, cmd = m.modelHostList.Update(msg)
		return m, cmd
	}

	var cmd tea.Cmd
	m.modelSSHSession, cmd = m.modelSSHSession.Update(msg)
	return m, cmd
	}*/
	return m, nil
}

// func (m *MainModel) embeddedSessionUsesHostList() bool {
// 	return m.appState.CurrentView == state.ViewSSHSession && m.appState.EmbeddedTerminalEnabled
// }

/*
func (m *Model) handleSessionSplitMouse(msg tea.MouseMsg) tea.Cmd {
	listWidth, _ := sessionPaneWidths(m.appState.Width)
	if msg.Mouse().X >= listWidth {
		return nil
	}

	var cmd tea.Cmd
	m.modelHostList, cmd = m.modelHostList.Update(msg)
	return cmd
}

func (m *Model) setSessionSplit(open bool) tea.Cmd {
	m.sessionSplit = open
	m.syncSessionSidebar()
	if session, ok := m.modelSSHSession.(*sshsession.Model); ok && session != nil {
		if open {
			session.Blur()
		} else {
			session.Focus()
		}
	}

	return m.layoutSessionPanes()
}

func (m *Model) syncSessionSidebar() {
	list, ok := m.modelHostList.(*hostlist.Model)
	if !ok || list == nil {
		return
	}

	list.SetSessionSidebar(m.sessionSplit)
}

func (m *Model) clearSessionSplit() {
	if !m.sessionSplit {
		return
	}

	m.sessionSplit = false
	m.syncSessionSidebar()
	m.modelHostList, _ = m.resizeChild(m.modelHostList, m.appState.Width, m.appState.Height)
}

func (m *Model) closeSSHSession() {
	session, ok := m.modelSSHSession.(*sshsession.Model)
	if !ok || session == nil {
		return
	}

	session.Close()
}

func (m *Model) layoutSessionPanes() tea.Cmd {
	listWidth := m.appState.Width
	sessionWidth := m.appState.Width
	if m.sessionSplit {
		listWidth, sessionWidth = sessionPaneWidths(m.appState.Width)
	}

	var cmd tea.Cmd
	var cmds []tea.Cmd
	m.modelHostList, cmd = m.resizeChild(m.modelHostList, listWidth, m.appState.Height)
	cmds = append(cmds, cmd)
	m.modelSSHSession, cmd = m.resizeChild(m.modelSSHSession, sessionWidth, m.appState.Height)
	cmds = append(cmds, cmd)
	m.modelGroupList, cmd = m.resizeChild(m.modelGroupList, m.appState.Width, m.appState.Height)
	cmds = append(cmds, cmd)

	return tea.Batch(cmds...)
}

func (m *Model) resizeChild(child tea.Model, width, height int) (tea.Model, tea.Cmd) {
	if child == nil || width <= 0 || height <= 0 {
		return child, nil
	}

	return child.Update(tea.WindowSizeMsg{Width: width, Height: height})
} */

/*
func (m *Model) sessionSplitView() tea.View {
	listWidth, _ := sessionPaneWidths(m.appState.Width)
	left := lipgloss.NewStyle().Width(listWidth).Render(m.modelHostList.View().Content)
	divider := sessionSplitDivider(max(m.appState.Height, 1))
	right := ""
	if m.modelSSHSession != nil {
		right = m.modelSSHSession.View().Content
	}

	return tea.NewView(lipgloss.JoinHorizontal(lipgloss.Top, left, divider, right))
}

func sessionSplitDivider(height int) string {
	if height < 1 {
		return "│"
	}

	return strings.TrimRight(strings.Repeat("│\n", height), "\n")
}
*/
