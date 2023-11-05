//go:build !windows

package ui

import tea "github.com/charmbracelet/bubbletea"

func (m *mainModel) Init() tea.Cmd {
	m.logger.Debug("Linux version")
	m.activeComponent = m.modelHostList
	initCmd := m.activeComponent.Init()

	// switch m.state {
	// case viewEditItem:
	// 	m.modelEditHost.Init()
	// default:
	// 	m.modelHostList.Init()
	// }

	return initCmd
}
