package sshcommand

import (
	"strings"

	"github.com/grafviktor/goto/internal/model/sshconfig"
)

var baseCmd = BaseCMD()

// ConnectCommand - builds ssh command to connect to a remote host.
func ConnectCommand(options ...Option) string {
	sb := strings.Builder{}
	sb.WriteString(baseCmd)

	for _, option := range options {
		addOption(&sb, option)
	}

	if sshconfig.IsUserDefinedPath() {
		addOption(&sb, OptionConfigFilePath{Value: sshconfig.Path()})
	}

	return sb.String()
}

// LoadConfigCommand - builds ssh command to load config from ssh_config file.
func LoadConfigCommand(options ...Option) string {
	sb := strings.Builder{}
	sb.WriteString(baseCmd)

	for _, option := range options {
		addOption(&sb, option)
	}

	if sshconfig.IsUserDefinedPath() {
		addOption(&sb, OptionConfigFilePath{Value: sshconfig.Path()})
	}

	return sb.String()
}
