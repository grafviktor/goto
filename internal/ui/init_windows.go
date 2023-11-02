//go:build windows

package ui

import (
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"golang.org/x/term"
)

var pollingPeriod = 1 * time.Second

func (m mainModel) pollWindowSize() {
	terminalFd := int(os.Stdin.Fd())
	w, h, _ := term.GetSize(terminalFd)
	for {
		time.Sleep(pollingPeriod)
		newW, newH, _ := term.GetSize(terminalFd)
		if newW != w || newH != h {
			w = newW
			h = newH
			m.Update(tea.WindowSizeMsg{w, h})
		}
	}
}

func (m mainModel) Init() tea.Cmd {
	switch m.state {
	case viewEditItem:
		return m.modelEditHost.Init()
	case viewHostList:
		return m.modelHostList.Init()
	}

	go m.pollWindowSize()

	return nil
}
