// Package host contains definition of Host data model.
package host

import (
	"strings"

	"github.com/grafviktor/goto/internal/constant"
	"github.com/grafviktor/goto/internal/model/sshcommand"
	"github.com/grafviktor/goto/internal/model/sshconfig"
)

// NewHost - constructs new Host model.
func NewHost(id int, title, description, address, loginName, identityFilePath, remotePort string) Host {
	return Host{
		ID:               id,
		Title:            title,
		Description:      description,
		Address:          address,
		LoginName:        loginName,
		RemotePort:       remotePort,
		IdentityFilePath: identityFilePath,
	}
}

// Host model definition.
type Host struct {
	ID               int                      `yaml:"-"`
	Title            string                   `yaml:"title"`
	Description      string                   `yaml:"description,omitempty"`
	Group            string                   `yaml:"group,omitempty"`
	Address          string                   `yaml:"address"`
	RemotePort       string                   `yaml:"network_port,omitempty"`
	LoginName        string                   `yaml:"username,omitempty"`
	IdentityFilePath string                   `yaml:"identity_file_path,omitempty"`
	SSHClientConfig  *sshconfig.Config        `yaml:"-"` // TODO: Must be renamed to SSHHostConfig
	StorageType      constant.HostStorageEnum `yaml:"-"`
}

// Clone host model.
func (h *Host) Clone() Host {
	newHost := Host{
		Title:            h.Title,
		Group:            h.Group,
		Description:      h.Description,
		Address:          h.Address,
		LoginName:        h.LoginName,
		IdentityFilePath: h.IdentityFilePath,
		RemotePort:       h.RemotePort,
	}

	return newHost
}

// IsUserDefinedSSHCommand returns true if the address contains spaces or "@" symbol,
// true means that user uses a custom config and not relying on LoginName, IdentityFilePath
// and RemotePort.
func (h *Host) IsUserDefinedSSHCommand() bool {
	rawValue := strings.TrimSpace(h.Address)
	containsSpace := strings.Contains(rawValue, " ")
	containsAtSymbol := strings.Contains(rawValue, "@")

	return containsSpace || containsAtSymbol
}

// CmdSSHConnect - returns SSH command for connecting to a remote host.
func (h *Host) CmdSSHConnect() string {
	if h.IsUserDefinedSSHCommand() {
		return sshcommand.ConnectCommand(sshcommand.OptionAddress{Value: h.Address})
	}

	if h.StorageType == constant.HostStorageType.SSH_CONFIG {
		// When it's SSH_CONFIG storage type, we need to use the title as a host name.
		// This is because the by addressing the host by alias, we get all its settings from ssh_config.
		return sshcommand.ConnectCommand(sshcommand.OptionAddress{Value: h.Title})
	}

	return sshcommand.ConnectCommand([]sshcommand.Option{
		sshcommand.OptionPrivateKey{Value: h.IdentityFilePath},
		sshcommand.OptionRemotePort{Value: h.RemotePort},
		sshcommand.OptionLoginName{Value: h.LoginName},
		sshcommand.OptionAddress{Value: h.Address},
	}...)
}

// CmdSSHConfig - returns SSH command for loading host default configuration.
func (h *Host) CmdSSHConfig() string {
	if h.StorageType == constant.HostStorageType.SSH_CONFIG {
		return sshcommand.LoadConfigCommand(sshcommand.OptionReadHostConfig{Value: h.Title})
	}

	if h.IsUserDefinedSSHCommand() {
		return sshcommand.LoadConfigCommand(sshcommand.OptionReadHostConfig{Value: h.Address})
	}

	return sshcommand.LoadConfigCommand([]sshcommand.Option{
		sshcommand.OptionPrivateKey{Value: h.IdentityFilePath},
		sshcommand.OptionRemotePort{Value: h.RemotePort},
		sshcommand.OptionLoginName{Value: h.LoginName},
		sshcommand.OptionReadHostConfig{Value: h.Address},
	}...)
}

// CmdSSHCopyID - returns SSH command for copying SSH key to a remote host (see ssh-copy-id).
func (h *Host) CmdSSHCopyID() string {
	// Though ssh-copy-id respects ssh_config, it's impossible to specify alternative ssh config file.
	if h.StorageType == constant.HostStorageType.SSH_CONFIG {
		return sshcommand.CopyIDCommand(sshcommand.OptionAddress{Value: h.Title})
	}

	return sshcommand.CopyIDCommand(
		sshcommand.OptionLoginName{Value: h.SSHClientConfig.User},
		sshcommand.OptionRemotePort{Value: h.SSHClientConfig.Port},
		sshcommand.OptionPrivateKey{Value: h.SSHClientConfig.IdentityFile},
		sshcommand.OptionAddress{Value: h.SSHClientConfig.Hostname},
	)
}
