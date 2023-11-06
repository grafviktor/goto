//go:build windows

package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/grafviktor/goto/internal/ui/message"
)

// var pollingPeriod = time.Second / 4

// func (m *mainModel) pollWindowSize() {
// 	m.logger.Debug("pollWindowSize()")
// 	terminalFd := int(os.Stdout.Fd())
// 	w, h, _ := term.GetSize(terminalFd)
// 	m.Update(tea.WindowSizeMsg{Width: w, Height: h})
// 	m.logger.Debug("Terminal window current size: %d %d", w, h)
// 	for {
// 		time.Sleep(pollingPeriod)
// 		newW, newH, _ := term.GetSize(terminalFd)
// 		if newW != w || newH != h {
// 			w = newW
// 			h = newH
// 			// m.logger.Debug("Terminal window new size: %d %d", w, h)
// 			m.Update(tea.WindowSizeMsg{Width: w, Height: h})
// 		}
// 	}
// }

func (m *mainModel) Init() tea.Cmd {
	m.logger.Debug("Windows version")
	// go func() {
	// 	m.logger.Debug("Polling terminal size every %d seconds", pollingPeriod/time.Second)
	// 	m.pollWindowSize()
	// }()
	cmd := m.modelHostList.Init()

	// switch m.state {
	// case viewEditItem:
	// 	m.modelEditHost.Init()
	// default:

	// }

	return tea.Batch(cmd, message.Tick)
}
