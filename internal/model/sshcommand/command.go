package sshcommand

import (
	"strings"

	"github.com/grafviktor/goto/internal/model/sshconfig"
)

var baseCmd = BaseCMD()

// Build - builds ssh command to connect to a remote host or load config from ssh_config file.
func Build(options ...Option) string {
	sb := strings.Builder{}
	sb.WriteString(baseCmd)

	for _, option := range options {
		addOption(&sb, option)
	}

	if sshconfig.IsEnabled() && sshconfig.IsUserDefinedPath() {
		addOption(&sb, OptionConfigFilePath{Value: sshconfig.Path()})
	}

	return sb.String()
}
