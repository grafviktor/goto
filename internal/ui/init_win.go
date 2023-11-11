//go:build windows

package ui

import (
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"golang.org/x/term"

	"github.com/grafviktor/goto/internal/ui/message"
)

var terminalSizePollingInterval = time.Second / 2

func TerminalSizePolling() tea.Msg {
	time.Sleep(terminalSizePollingInterval)
	terminalFd := int(os.Stdout.Fd())
	Width, Height, _ := term.GetSize(terminalFd)

	return message.TerminalSizePollingMsg{Width, Height}
}

func (m *mainModel) Init() tea.Cmd {
	m.logger.Debug("Run Windows OS specific UI init function")
	cmd := m.modelHostList.Init()

	m.logger.Debug("Start polling terminal size")
	return tea.Batch(cmd, message.TerminalSizePolling)
}
