package ui

import (
	"context"
	"os"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/grafviktor/goto/internal/state"
	"github.com/grafviktor/goto/internal/storage"
	"github.com/grafviktor/goto/internal/ui/component/edithost"
	"github.com/grafviktor/goto/internal/ui/component/hostlist"
	"github.com/grafviktor/goto/internal/ui/message"
	"golang.org/x/term"
)

type sessionState int

const (
	viewHostList sessionState = iota
	viewEditItem
)

type logger interface {
	Debug(format string, args ...any)
}

func NewMainModel(ctx context.Context, storage storage.HostStorage, appState *state.ApplicationState, log logger) mainModel {
	m := mainModel{
		modelHostList: hostlist.New(ctx, storage, appState, log),
		appContext:    ctx,
		hostStorage:   storage,
		appState:      appState,
		logger:        log,
	}

	return m
}

type mainModel struct {
	appContext      context.Context
	hostStorage     storage.HostStorage
	state           sessionState
	activeComponent tea.Model
	modelHostList   tea.Model
	modelEditHost   tea.Model
	// TODO: Move mainModel to "State" object or vice versa
	appState *state.ApplicationState
	logger   logger
	viewport viewport.Model
	ready    bool
}

func (m *mainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		return m.handleKeyEvent(keyMsg)
	}

	switch msg := msg.(type) {
	case tickMsg:
		terminalFd := int(os.Stdout.Fd())
		w, h, _ := term.GetSize(terminalFd)
		if w != m.appState.Width || h != m.appState.Height {
			m.updateViewPort(w, h)
			cmds = append(cmds, message.TeaCmd(tea.WindowSizeMsg{Width: w, Height: h}))
		}
		cmds = append(cmds, tick)
	case tea.WindowSizeMsg:
		m.logger.Debug("Terminal window new size: %d %d", msg.Width, msg.Height)
		m.appState.Width = msg.Width
		m.appState.Height = msg.Height
		m.updateViewPort(msg.Width, msg.Height)
	case hostlist.MsgEditItem:
		// cmds = append(cmds, func() tea.Msg { return msg })
		m.state = viewEditItem
		ctx := context.WithValue(m.appContext, edithost.ItemID, msg.HostID)
		m.modelEditHost = edithost.New(ctx, m.hostStorage, m.appState, m.logger)
	case hostlist.MsgNewItem:
		m.state = viewEditItem
		m.modelEditHost = edithost.New(m.appContext, m.hostStorage, m.appState, m.logger)
	case hostlist.MsgSelectItem:
		m.appState.Selected = msg.HostID
	case edithost.MsgClose:
		m.state = viewHostList
	}

	m.modelHostList, cmd = m.modelHostList.Update(msg)
	cmds = append(cmds, cmd)

	if m.state == viewEditItem {
		// edit host receives messages only if it's active. We re-create this component every time we go to edit mode
		m.modelEditHost, cmd = m.modelEditHost.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *mainModel) handleKeyEvent(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.Type == tea.KeyCtrlC {
		return m, tea.Quit
	}

	var cmd tea.Cmd
	switch m.state { // Only active component receives key messages
	case viewHostList:
		m.modelHostList, cmd = m.modelHostList.Update(msg)
	case viewEditItem:
		m.modelEditHost, cmd = m.modelEditHost.Update(msg)
	}

	return m, cmd
}

func (m *mainModel) View() string {
	var content string
	switch m.state {
	case viewEditItem:
		content = m.modelEditHost.View()
	case viewHostList:
		content = m.modelHostList.View()
	}

	m.viewport.SetContent(content)

	return m.viewport.View()
}

func (m *mainModel) updateViewPort(w, h int) tea.Model {
	if !m.ready {
		m.ready = true
		m.viewport = viewport.New(m.appState.Width, m.appState.Height)
	} else {
		m.viewport.Width = w
		m.viewport.Height = h
	}

	return m
}
