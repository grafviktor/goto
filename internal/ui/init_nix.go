//go:build !windows

package ui

import tea "github.com/charmbracelet/bubbletea"

func (m *mainModel) Init() tea.Cmd {
	m.logger.Debug("[UI] Run init function")
	cmd := m.modelHostList.Init()

	return cmd
}
