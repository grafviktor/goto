// Package message contains shared messages which are used to communicate between bubbletea components
package message

import (
	"os"
	"time"

	"golang.org/x/term"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/grafviktor/goto/internal/constant"
	"github.com/grafviktor/goto/internal/model/host"
	"github.com/grafviktor/goto/internal/model/ssh"
)

type (
	// InitComplete - is a message which is sent when bubbletea models are initialized.
	InitComplete struct{}
	// TerminalSizePolling - is a message which is sent when terminal width and/or height changes.
	TerminalSizePolling struct{ Width, Height int }
	// HostListSelectItem is required to let host list know that it's time to update title.
	HostListSelectItem struct{ HostID int }
	// HostCreated - is dispatched when a new host was added to the database.
	HostCreated struct{ Host host.Host }
	// HostUpdated - is dispatched when host model is updated.
	HostUpdated struct{ Host host.Host }
	// HostSSHConfigLoaded triggers when app loads a host config using ssh -G <hostname>.
	// The config is stored in main model: m.appState.HostSSHConfig.
	HostSSHConfigLoaded struct {
		HostID int
		Config ssh.Config
	}
	// Open available host groups form
	OpenSelectGroupForm struct{}
	// Close available host groups form
	CloseSelectGroupForm struct{}
	// GroupListSelectItem - is dispatched when select a group in grouplist.
	GroupListSelectItem struct{ GroupName string }
	// OpenEditForm fires when user press edit button on a selected host
	OpenEditForm struct{ HostID int }
	// CloseEditForm triggers when users exits from edit form without saving results.
	CloseEditForm struct{}
	// ErrorOccurred - is dispatched when an error occurs.
	ErrorOccurred struct{ Err error }
	// RunProcessSSHConnect is dispatched when user wants to connect to a host.
	RunProcessSSHConnect struct{ Host host.Host }
	// RunProcessSSHLoadConfig is dispatched it's required to read .ssh/config file for a certain host.
	RunProcessSSHLoadConfig struct{ Host host.Host }
	// RunProcessSSHCopyID is dispatched when user wants to copy SSH key to a remote host.
	RunProcessSSHCopyID struct{ Host host.Host }
	// RunProcessErrorOccurred fires when there is an error executing an external process.
	RunProcessErrorOccurred struct {
		ProcessType constant.ProcessType
		StdOut      string // Even if process fails, it may have some output.
		StdErr      string
	}
	// RunProcessSuccess fires when external process exits normally.
	RunProcessSuccess struct {
		ProcessType constant.ProcessType
		StdOut      string
		StdErr      string // Even if process succeeds, it may have some output.
	}
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
