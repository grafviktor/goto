// Package host contains definition of Host data model.
package host

import (
	"strings"

	"github.com/grafviktor/goto/internal/constant"
	"github.com/grafviktor/goto/internal/model/ssh"
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
	SSHClientConfig  *ssh.Config              `yaml:"-"`
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
		return ssh.ConnectCommand(ssh.OptionAddress{Value: h.Address})
	}

	return ssh.ConnectCommand([]ssh.Option{
		ssh.OptionPrivateKey{Value: h.IdentityFilePath},
		ssh.OptionRemotePort{Value: h.RemotePort},
		ssh.OptionLoginName{Value: h.LoginName},
		ssh.OptionAddress{Value: h.Address},
	}...)
}

// CmdSSHConfig - returns SSH command for loading host default configuration.
func (h *Host) CmdSSHConfig() string {
	if h.IsUserDefinedSSHCommand() {
		return ssh.LoadConfigCommand(ssh.OptionReadConfig{Value: h.Address})
	}

	return ssh.LoadConfigCommand([]ssh.Option{
		ssh.OptionPrivateKey{Value: h.IdentityFilePath},
		ssh.OptionRemotePort{Value: h.RemotePort},
		ssh.OptionLoginName{Value: h.LoginName},
		ssh.OptionReadConfig{Value: h.Address},
	}...)
}

// CmdSSHCopyID - returns SSH command for copying SSH key to a remote host (see ssh-copy-id).
func (h *Host) CmdSSHCopyID() string {
	hostname := h.SSHClientConfig.Hostname
	identityFile := h.SSHClientConfig.IdentityFile
	port := h.SSHClientConfig.Port
	user := h.SSHClientConfig.User

	return ssh.CopyIDCommand(
		ssh.OptionLoginName{Value: user},
		ssh.OptionRemotePort{Value: port},
		ssh.OptionPrivateKey{Value: identityFile},
		ssh.OptionAddress{Value: hostname},
	)
}
