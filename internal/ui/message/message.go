package message

import (
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"golang.org/x/term"
)

type (
	InitComplete           struct{}
	TerminalSizePollingMsg struct{ Width, Height int }
)

var terminalSizePollingInterval = time.Second / 2

func TerminalSizePolling() tea.Msg {
	time.Sleep(terminalSizePollingInterval)
	terminalFd := int(os.Stdout.Fd())
	Width, Height, _ := term.GetSize(terminalFd)

	return TerminalSizePollingMsg{Width, Height}
}

// A helper function which returns create tea.Cmd from tea.Msg object
func TeaCmd(msg any) func() tea.Msg {
	return func() tea.Msg {
		return msg
	}
}
