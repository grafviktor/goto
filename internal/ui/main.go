// Package ui - contains UI iteraction code.
package ui

import (
	"context"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/grafviktor/goto/internal/state"
	"github.com/grafviktor/goto/internal/storage"
	"github.com/grafviktor/goto/internal/ui/component/edithost"
	"github.com/grafviktor/goto/internal/ui/component/hostlist"
	"github.com/grafviktor/goto/internal/ui/message"
)

type logger interface {
	Debug(format string, args ...any)
	Error(format string, args ...any)
}

// NewMainModel - creates a parent module for other component and preserves stat which
// can be propagated to other sub-components.
func NewMainModel(
	ctx context.Context,
	storage storage.HostStorage,
	appState *state.ApplicationState,
	log logger,
) mainModel {
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
	appContext    context.Context
	hostStorage   storage.HostStorage
	modelHostList tea.Model
	modelEditHost tea.Model
	appState      *state.ApplicationState
	logger        logger
	viewport      viewport.Model
	ready         bool
}

func (m *mainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		m.logger.Debug("[UI] Keyboard event: %v", msg)
		return m.handleKeyEvent(keyMsg)
	}

	switch msg := msg.(type) {
	case message.TerminalSizePolling:
		// That is Windows OS specific. Windows cmd.exe does not trigger terminal
		// resize events, that is why we poll terminal size with intervals
		// First message is being triggered by Windows version of the model.Init function.
		if msg.Width != m.appState.Width || msg.Height != m.appState.Height {
			m.logger.Debug("[UI] Terminal size polling message received: %d %d", msg.Width, msg.Height)
			cmds = append(cmds, message.TeaCmd(tea.WindowSizeMsg{Width: msg.Width, Height: msg.Height}))
		}

		// We're dispatching the same message from this function and therefore cycling TerminalSizePollingMsg.
		// That's done on purpose to keep this process running. Message.TerminalSizePollingMsg will trigger
		// automatically after an artificial delay which set by Time.Sleep inside message.
		cmds = append(cmds, message.TerminalSizePollingMsg)
	case tea.WindowSizeMsg:
		m.logger.Debug("[UI] Set terminal window size: %d %d", msg.Width, msg.Height)
		m.appState.Width = msg.Width
		m.appState.Height = msg.Height
		m.updateViewPort(msg.Width, msg.Height)
	case hostlist.MsgEditItem:
		m.logger.Debug("[UI] Open host edit form")
		m.appState.CurrentView = state.ViewEditItem
		ctx := context.WithValue(m.appContext, edithost.ItemID, msg.HostID)
		m.modelEditHost = edithost.New(ctx, m.hostStorage, m.appState, m.logger)
	case hostlist.MsgNewItem:
		m.logger.Debug("[UI] Create a new host")
		m.appState.CurrentView = state.ViewEditItem
		m.modelEditHost = edithost.New(m.appContext, m.hostStorage, m.appState, m.logger)
	case message.HostListSelectItem:
		m.logger.Debug("[UI] Select host id: %d", msg.HostID)
		m.appState.Selected = msg.HostID
	case edithost.MsgClose:
		m.logger.Debug("[UI] Close host edit form")
		m.appState.CurrentView = state.ViewHostList
	case message.RunProcessErrorOccured:
		m.logger.Debug("[UI] External process error. %v", msg.Err)
		m.appState.Err = msg.Err
		m.appState.CurrentView = state.ViewErrorMessage
	}

	m.modelHostList, cmd = m.modelHostList.Update(msg)
	cmds = append(cmds, cmd)

	if m.appState.CurrentView == state.ViewEditItem {
		// Edit host receives messages only if it's active. We re-create this component every time we go to edit mode
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

	// Only current view receives key messages
	switch m.appState.CurrentView {
	case state.ViewErrorMessage:
		// When display external process's output and receieve any keyboard event, we:
		// 1. Reset the error message
		// 2. Switch to HostList view
		m.appState.Err = nil
		m.appState.CurrentView = state.ViewHostList
	case state.ViewHostList:
		m.modelHostList, cmd = m.modelHostList.Update(msg)
	case state.ViewEditItem:
		m.modelEditHost, cmd = m.modelEditHost.Update(msg)
	}

	return m, cmd
}

func (m *mainModel) View() string {
	// Build UI
	var content string
	switch m.appState.CurrentView {
	case state.ViewErrorMessage:
		content = m.appState.Err.Error()
	case state.ViewEditItem:
		content = m.modelEditHost.View()
	case state.ViewHostList:
		content = m.modelHostList.View()
	}

	// Wrap UI into the ViewPort
	m.viewport.SetContent(content)
	viewPortContent := m.viewport.View()

	return viewPortContent
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
