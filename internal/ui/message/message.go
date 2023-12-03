// Package message contains shared messages which are used to communicate between bubbletea components
package message

import (
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"golang.org/x/term"
)

type (
	// InitComplete - is a message which is sent when bubbletea models are initialized.
	InitComplete struct{}
	// TerminalSizePolling - is a message which is sent when terminal width and/or height changes.
	TerminalSizePolling struct{ Width, Height int }
	// RemoteSessionErrorOccured fires when there is an error connecting to a remote host.
	RemoteSessionErrorOccured struct{ Err error }
)

var terminalSizePollingInterval = time.Second / 2

// TerminalSizePollingMsg - is a tea.Msg which is used to poll terminal size.
func TerminalSizePollingMsg() tea.Msg {
	time.Sleep(terminalSizePollingInterval)
	terminalFd := int(os.Stdout.Fd())
	Width, Height, _ := term.GetSize(terminalFd)

	return TerminalSizePolling{Width, Height}
}

// TeaCmd - is a helper function which returns create tea.Cmd from tea.Msg object.
func TeaCmd(msg any) func() tea.Msg {
	return func() tea.Msg {
		return msg
	}
}
