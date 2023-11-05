//go:build windows

package ui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
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

type tickMsg struct{}

func tick() tea.Msg {
	time.Sleep(time.Second * 4)
	return tickMsg{}
}

func (m *mainModel) Init() tea.Cmd {
	m.logger.Debug("Windows version")
	// go func() {
	// 	m.logger.Debug("Polling terminal size every %d seconds", pollingPeriod/time.Second)
	// 	m.pollWindowSize()
	// }()

	switch m.state {
	case viewEditItem:
		m.modelEditHost.Init()
	default:
		m.modelHostList.Init()
	}

	return tick
}
