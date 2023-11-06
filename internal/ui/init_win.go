//go:build windows

package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/grafviktor/goto/internal/ui/message"
)

func (m *mainModel) Init() tea.Cmd {
	m.logger.Debug("Windows version")
	cmd := m.modelHostList.Init()

	m.logger.Debug("Start polling terminal size")
	return tea.Batch(cmd, message.TerminalSizePolling)
}
