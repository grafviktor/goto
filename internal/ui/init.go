//go:build !windows

package ui

import tea "github.com/charmbracelet/bubbletea"

func (m mainModel) Init() tea.Cmd {
	switch m.state {
	case viewEditItem:
		return m.modelEditHost.Init()
	case viewHostList:
		return m.modelHostList.Init()
	}

	return nil
}
