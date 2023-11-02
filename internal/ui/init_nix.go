//go:build !windows

package ui

import tea "github.com/charmbracelet/bubbletea"

func (m mainModel) Init() tea.Cmd {
	m.logger.Debug("Linux version")
	switch m.state {
	case viewEditItem:
		return m.modelEditHost.Init()
	case viewHostList:
		return m.modelHostList.Init()
	}

	return nil
}
