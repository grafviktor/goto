// Package message contains shared messages which are used to communicate between bubbletea components
package message

import (
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"golang.org/x/term"

	"github.com/grafviktor/goto/internal/model"
)

type (
	// InitComplete - is a message which is sent when bubbletea models are initialized.
	InitComplete struct{}
	// TerminalSizePolling - is a message which is sent when terminal width and/or height changes.
	TerminalSizePolling struct{ Width, Height int }
	// RunProcessErrorOccurred fires when there is an error executing an external process.
	RunProcessErrorOccurred struct{ Err error }
	// HostListSelectItem is required to let host list know that it's time to update title.
	HostListSelectItem struct{ HostID int }
	// RunProcessConnectSSH is dispatched when user wants to connect to a host.
	RunProcessConnectSSH struct{ Host model.Host }
	// RunProcessLoadSSHConfig is dispatched it's required to read .ssh/config file for a certain host.
	RunProcessLoadSSHConfig struct{ SSHConfigHostname string }
)

var terminalSizePollingInterval = time.Second / 2

// TerminalSizePollingMsg - is a tea.Msg which is used to poll terminal size.
func TerminalSizePollingMsg() tea.Msg {
	time.Sleep(terminalSizePollingInterval)
	terminalFd := int(os.Stdout.Fd())
	Width, Height, _ := term.GetSize(terminalFd)

	return TerminalSizePolling{Width, Height}
}

// TeaCmd - is a helper function which creates tea.Cmd from tea.Msg object.
func TeaCmd(msg any) func() tea.Msg {
	return func() tea.Msg {
		return msg
	}
}
