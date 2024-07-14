// Package constant contains shared app constants
package constant

import "errors"

// ErrNotFound is used by data layer.
var ErrNotFound = errors.New("not found")

// ProtocolSSH - is only supported protocol.
const ProtocolSSH = "ssh"

// ScreenLayout is used to determine how the hostlist should be displayed.
type ScreenLayout string

const (
	// ScreenLayoutTight is set when all items in the hostlist are shown without description.
	ScreenLayoutTight ScreenLayout = "tight"
	// ScreenLayoutNormal is set when all hosts are shown with description field and a margin.
	ScreenLayoutNormal ScreenLayout = "normal"
)

type ProcessType string

const (
	ProcessTypeSSHLoadConfig ProcessType = "ssh-load-config"
	ProcessTypeSSHCopyID     ProcessType = "ssh-copy-id"
	ProcessTypeSSHConnect    ProcessType = "ssh-connect"
)
