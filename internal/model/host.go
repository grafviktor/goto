// Package model contains description of data models. For now there is only 'Host' model
package model

import (
	"fmt"
	"strings"
)

var baseCmd = BaseCMD()

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

// CmdSSHConnect - returns SSH command for connecting to a remote host
func (h *Host) CmdSSHConnect() string {
	if h.IsUserDefinedSSHCommand() {
		return fmt.Sprintf("%s %s", baseCmd, h.Address)
	}

	sb := strings.Builder{}
	sb.WriteString(baseCmd)

	options := hostModelToOptionsAdaptor(*h)
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
