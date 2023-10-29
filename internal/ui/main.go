package ui

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/grafviktor/goto/internal/state"
	"github.com/grafviktor/goto/internal/storage"
	"github.com/grafviktor/goto/internal/ui/component/edithost"
	"github.com/grafviktor/goto/internal/ui/component/hostlist"
)

type sessionState int

const (
	viewHostList sessionState = iota
	viewEditItem
)

func NewMainModel(ctx context.Context, storage storage.HostStorage, appState *state.ApplicationState) mainModel {
	m := mainModel{
		modelHostList: hostlist.New(ctx, storage, appState),
		appContext:    ctx,
		hostStorage:   storage,
		appState:      appState,
	}

	return m
}

type mainModel struct {
	appContext    context.Context
	hostStorage   storage.HostStorage
	state         sessionState
	modelHostList tea.Model
	modelEditHost tea.Model
	// TODO: Move mainModel to "State" object or vice versa
	appState *state.ApplicationState
}

func (m mainModel) Init() tea.Cmd {
	switch m.state {
	case viewEditItem:
		return m.modelEditHost.Init()
	case viewHostList:
		return m.modelHostList.Init()
	}

	return nil
}

func (m mainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		return m.handleKeyEvent(keyMsg)
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.appState.Width = msg.Width
		m.appState.Height = msg.Height
	case hostlist.MsgEditItem:
		m.state = viewEditItem
		ctx := context.WithValue(m.appContext, edithost.ItemID, msg.HostID)
		m.modelEditHost = edithost.New(ctx, m.hostStorage, m.appState)
	case hostlist.MsgNewItem:
		m.state = viewEditItem
		m.modelEditHost = edithost.New(m.appContext, m.hostStorage, m.appState)
	case hostlist.MsgSelectItem:
		m.appState.Selected = msg.Index
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

	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m mainModel) handleKeyEvent(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.Type == tea.KeyCtrlC {
		// TODO: Save State
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

func (m mainModel) View() string {
	switch m.state {
	case viewEditItem:
		return m.modelEditHost.View()
	case viewHostList:
		return m.modelHostList.View()
	}

	panic("Should not be here")
}
