// Package message contains shared messages which are used to communicate between bubbletea components
package message

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/grafviktor/goto/internal/constant"
	"github.com/grafviktor/goto/internal/model/host"
	"github.com/grafviktor/goto/internal/model/sshconfig"
)

type (
	// InitComplete - is a message which is sent when bubbletea models are initialized.
	InitComplete struct{}
	// TerminalSizePolling - is a message which is sent when terminal width and/or height changes.
	TerminalSizePolling struct{ Width, Height int }
	// HostSelected is required to let host list know that it's time to update title.
	HostSelected struct{ HostID int }
	// HostCreated - is dispatched when a new host was added to the database.
	HostCreated struct{ Host host.Host }
	// HostUpdated - is dispatched when host model is updated.
	HostUpdated struct{ Host host.Host }
	// HostSSHConfigLoaded triggers when app loads a host config using ssh -G <hostname>.
	// The config is stored in main model: m.appState.HostSSHConfig.
	HostSSHConfigLoaded struct {
		HostID int
		Config sshconfig.Config
	}
	// OpenViewSelectGroup - dispatched when it's required to open group list view.
	OpenViewSelectGroup struct{}
	// CloseViewSelectGroup - dispatched when it's required to close group list view.
	CloseViewSelectGroup struct{}
	// GroupSelected - is dispatched when select a group in group list view.
	GroupSelected struct{ Name string }
	// HideUINotification - is dispatched when it's time to hide UI notification and display normal component's title.
	HideUINotification struct{ ComponentName string }
	// OpenViewHostEdit fires when user press edit button on a selected host.
	OpenViewHostEdit struct{ HostID int }
	// CloseViewHostEdit triggers when users exits from edit form without saving results.
	CloseViewHostEdit struct{}
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

// TeaCmd - is a helper function which creates tea.Cmd from tea.Msg object.
func TeaCmd(msg any) func() tea.Msg {
	return func() tea.Msg {
		return msg
	}
}

type titledUIModel interface {
	SetTitle(title string)
}

var (
	timers                         map[string]*time.Timer
	notificationMessageDisplayTime = time.Second * 2
)

// DisplayNotification - dispatches UI event to display a notification in the UI for a specific component.
func DisplayNotification(targetComponentName, text string, titledModel titledUIModel) tea.Cmd {
	if timers == nil {
		timers = make(map[string]*time.Timer)
	}

	if timer, ok := timers[targetComponentName]; ok {
		timer.Stop()
	}

	titledModel.SetTitle(text)
	timers[targetComponentName] = time.NewTimer(notificationMessageDisplayTime)

	return func() tea.Msg {
		<-timers[targetComponentName].C
		return HideUINotification{ComponentName: targetComponentName}
	}
}
