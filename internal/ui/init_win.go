//go:build windows

package ui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type tickMsg struct{}

func tick() tea.Msg {
	time.Sleep(time.Second * 4)
	return tickMsg{}
}

func (m *mainModel) Init() tea.Cmd {
	m.logger.Debug("Windows version")
	m.activeComponent = &m.modelHostList
	initCmd := m.modelHostList.Init()

	return tea.Batch(initCmd, tick)
}
