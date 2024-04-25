// Package model contains description of data models. For now there is only 'Host' model
package model

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/grafviktor/goto/internal/model/sshconfig"
	"github.com/grafviktor/goto/internal/utils"
)

var baseCmd = sshconfig.BaseCMD()

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
	ID               int    `yaml:"-"`
	Title            string `yaml:"title"`
	Description      string `yaml:"description,omitempty"`
	Address          string `yaml:"address"`
	RemotePort       string `yaml:"network_port,omitempty"`
	LoginName        string `yaml:"username,omitempty"`
	IdentityFilePath string `yaml:"identity_file_path,omitempty"`
	DefaultSSHConfig           sshconfig.Config `yaml:"-"`
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

func (h *Host) IsUserDefinedSSHCommand() bool {
	rawValue := strings.TrimSpace(h.Address)
	containsSpace := strings.Contains(rawValue, " ")
	containsAtSymbol := strings.Contains(rawValue, "@")

	return containsSpace || containsAtSymbol
}

// toSSHOptions - extract values from model.Host into a set of ssh.CommandLineOption
// host - model.Host to be adapted
// returns []ssh.CommandLineOption.
func (host *Host) toSSHOptions() []CommandLineOption {
	return []CommandLineOption{
		OptionPrivateKey{Value: host.IdentityFilePath},
		OptionRemotePort{Value: host.RemotePort},
		OptionLoginName{Value: host.LoginName},
		OptionAddress{Value: host.Address},
	}
}

// CmdSSHConnect - returns SSH command for connecting to a remote host
func (h *Host) CmdSSHConnect() string {
	if h.IsUserDefinedSSHCommand() {
		return fmt.Sprintf("%s %s", baseCmd, h.Address)
	}

	sb := strings.Builder{}
	sb.WriteString(baseCmd)

	options := h.toSSHOptions()
	for _, argument := range options {
		addOption(&sb, argument)
	}

	return sb.String()
}

// CmdSSHConnect - returns SSH command for connecting to a remote host
func (h *Host) CmdSSHConfig() string {
	sb := strings.Builder{}
	sb.WriteString(baseCmd)
	addOption(&sb, OptionReadConfig{Value: h.Address})

	return sb.String()
}

// BuildConnectSSH - builds ssh command which is based on host.Model.
func (h *Host) BuildConnectSSH() *exec.Cmd {
	command := h.CmdSSHConnect()
	process := utils.BuildProcess(command)
	process.Stdout = os.Stdout
	process.Stderr = &utils.ProcessBufferWriter{}

	return process
}

// BuildLoadSSHConfig - builds ssh command, which runs ssh -G <hostname> command
// to get a list of options associated with the hostname.
func (h *Host) BuildLoadSSHConfig() *exec.Cmd {
	// Use case 1: User edits host
	// Use case 2: User is going to copy his ssh key using <t> command from the hostlist

	command := h.CmdSSHConfig()
	process := utils.BuildProcess(command)
	process.Stdout = &utils.ProcessBufferWriter{}
	process.Stderr = &utils.ProcessBufferWriter{}

	return process
}