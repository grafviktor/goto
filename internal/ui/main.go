package ui

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/grafviktor/goto/internal/config"
	"github.com/grafviktor/goto/internal/storage"
	"github.com/grafviktor/goto/internal/ui/component/edithost"
	"github.com/grafviktor/goto/internal/ui/component/hostlist"
)

type sessionState int

const (
	viewHostList sessionState = iota
	viewEditItem
)

func NewMainModel(ctx context.Context, config config.Application, storage storage.HostStorage) mainModel {
	m := mainModel{
		modelHostList: hostlist.New(ctx, config, storage),
		appContext:    ctx,
		appConfig:     config,
		hostStorage:   storage,
	}

	return m
}

type mainModel struct {
	appContext    context.Context
	appConfig     config.Application
	hostStorage   storage.HostStorage
	state         sessionState
	modelHostList tea.Model
	modelEditHost tea.Model
	width         int
	height        int
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
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC {
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case hostlist.MsgEditItem:
		m.state = viewEditItem
		ctx := context.WithValue(m.appContext, edithost.ItemID, msg.HostID)
		m.modelEditHost = edithost.New(ctx, m.appConfig, m.hostStorage, m.width, m.height)
	case hostlist.MsgNewItem:
		m.state = viewEditItem
		m.modelEditHost = edithost.New(m.appContext, m.appConfig, m.hostStorage, m.width, m.height)
	case hostlist.MsgRepoUpdated:
		// HACK: хотя компонент hostList неактивен, мы отправляем ему сообщение вручную
		// наверное сообщения лучше передавать через каналы
		newHostList, _ := m.modelHostList.Update(msg)
		m.modelHostList = newHostList
	case edithost.MsgClose:
		m.state = viewHostList
	}

	var cmd tea.Cmd
	var cmds []tea.Cmd
	switch m.state {
	case viewHostList:
		newHostList, newCmd := m.modelHostList.Update(msg)
		m.modelHostList = newHostList
		cmd = newCmd
	case viewEditItem:
		newEditHost, newCmd := m.modelEditHost.Update(msg)
		m.modelEditHost = newEditHost
		cmd = newCmd
	}

	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
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
