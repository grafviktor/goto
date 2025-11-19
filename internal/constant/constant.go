// Package constant contains shared app constants
package constant

import "errors"

const (
	AppExitCodeSuccess = 0
	AppExitCodeError   = 1
)

// ErrNotFound is used by data layer.
var ErrNotFound = errors.New("not found")

// ScreenLayout is used to determine how the hostlist should be displayed.
type ScreenLayout string

const (
	// ScreenLayoutCompact is set when all items in the hostlist are shown without description.
	ScreenLayoutCompact ScreenLayout = "compact"
	// ScreenLayoutDescription is set when all hosts are shown with description field and a margin.
	ScreenLayoutDescription ScreenLayout = "description"
	// ScreenLayoutGroup is set when all hosts are shown with description field and a margin.
	ScreenLayoutGroup ScreenLayout = "group"
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

// HostStorageEnum defines the enum options for HostStorageType.
type HostStorageEnum string

// HostStorageType defines the type of the underlying storage of a host.
var HostStorageType = struct {
	Combined  HostStorageEnum
	SSHConfig HostStorageEnum
	YAMLFile  HostStorageEnum
}{
	Combined:  "COMBINED",
	SSHConfig: "SSH_CONFIG",
	YAMLFile:  "YAML_FILE",
}

type LogLevel = string

var LogLevelType = struct {
	INFO  LogLevel
	DEBUG LogLevel
}{
	INFO:  "info",
	DEBUG: "debug",
}

type AppMode = string

var AppModeType = struct {
	StartUI     AppMode
	DisplayInfo AppMode
	HandleParam AppMode
}{
	StartUI:     "START_UI",
	DisplayInfo: "DISPLAY_INFO",
	HandleParam: "HANDLE_PARAM",
}
