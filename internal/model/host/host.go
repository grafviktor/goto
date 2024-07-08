// Package host contains definition of Host data model.
package host

import (
	"strings"

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
	ID               int         `yaml:"-"`
	Title            string      `yaml:"title"`
	Description      string      `yaml:"description,omitempty"`
	Address          string      `yaml:"address"`
	RemotePort       string      `yaml:"network_port,omitempty"`
	LoginName        string      `yaml:"username,omitempty"`
	IdentityFilePath string      `yaml:"identity_file_path,omitempty"`
	DefaultSSHConfig *ssh.Config `yaml:"-"`
}

// Clone host model.
func (h *Host) Clone() Host {
	newHost := Host{
		Title:            h.Title,
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

// toSSHOptions - extract values from model.Host into a set of ssh.CommandLineOption
// host - model.Host to be adapted
// returns []ssh.CommandLineOption.
func (h *Host) toSSHOptions() []ssh.Option {
	return []ssh.Option{
		ssh.OptionPrivateKey{Value: h.IdentityFilePath},
		ssh.OptionRemotePort{Value: h.RemotePort},
		ssh.OptionLoginName{Value: h.LoginName},
		ssh.OptionAddress{Value: h.Address},
	}
}

// CmdSSHConnect - returns SSH command for connecting to a remote host.
func (h *Host) CmdSSHConnect() string {
	if h.IsUserDefinedSSHCommand() {
		return ssh.ConnectCommand(ssh.OptionAddress{Value: h.Address})
	}

	return ssh.ConnectCommand(h.toSSHOptions()...)
}

// CmdSSHConfig - returns SSH command for loading host default configuration.
func (h *Host) CmdSSHConfig() string {
	return ssh.LoadConfigCommand(ssh.OptionReadConfig{Value: h.Address})
}

func (h *Host) CmdSSHCopyID() string {
	user := h.DefaultSSHConfig.User
	port := h.DefaultSSHConfig.Port
	identityFile := h.DefaultSSHConfig.IdentityFile

	// FIXME: Should use address from the config struct as in the hostlist instead of a real hostname user can put ssh_config alias.
	// hostname := h.DefaultSSHConfig.Hostname

	return ssh.CopyIDCommand(
		ssh.OptionLoginName{Value: user},
		ssh.OptionRemotePort{Value: port},
		ssh.OptionPrivateKey{Value: identityFile},
		// ssh.OptionAddress{Value: hostname},
	)
}
