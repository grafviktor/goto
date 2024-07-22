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

// ProcessType is used to determine what kind of external process is running.
type ProcessType string

const (
	// ProcessTypeSSHLoadConfig is used when we need to run ssh -G <hostname> to get config.
	ProcessTypeSSHLoadConfig ProcessType = "ssh-load-config"
	// ProcessTypeSSHCopyID is used when we need to run ssh-copy-id to copy a public key to a remote host.
	ProcessTypeSSHCopyID ProcessType = "ssh-copy-id"
	// ProcessTypeSSHConnect is used when we want to connect to a remote host.
	ProcessTypeSSHConnect ProcessType = "ssh-connect"
)
