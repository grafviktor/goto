package ssh

import (
	"fmt"
	"strings"

	"github.com/grafviktor/goto/internal/utils"
)

// Option - parent interface for command line option.
type Option interface{}

type (
	// OptionPrivateKey - ssh private key path in file system.
	OptionPrivateKey struct{ Value string }
	// OptionRemotePort - Remote port to connect to.
	OptionRemotePort struct{ Value string }
	// OptionLoginName - is a login name which is used when connecting to a remote host. Ex: loginname@somehost.com.
	OptionLoginName struct{ Value string }
	// OptionAddress - is a remote host address. Example: somehost.com.
	OptionAddress struct{ Value string }
	// OptionReadConfig - is used to read config file from ssh_config. Cannot be combined with other options.
	OptionReadConfig struct{ Value string }
)

func constructKeyValueOption(optionFlag, optionValue string) string {
	optionValue = strings.TrimSpace(optionValue)
	if optionValue != "" {
		return fmt.Sprintf(" %s %s", optionFlag, optionValue)
	}
	return ""
}

func addOption(sb *strings.Builder, rawParameter Option) {
	var option string
	switch p := rawParameter.(type) {
	case OptionPrivateKey:
		option = constructKeyValueOption("-i", p.Value)
	case OptionRemotePort:
		option = constructKeyValueOption("-p", p.Value)
	case OptionLoginName:
		option = constructKeyValueOption("-l", p.Value)
	case OptionReadConfig:
		option = constructKeyValueOption("-G", utils.RemoveDuplicateSpaces(p.Value))
	case OptionAddress:
		if p.Value != "" {
			option = fmt.Sprintf(" %s", utils.RemoveDuplicateSpaces(p.Value))
		}
	default:
		return
	}

	sb.WriteString(option)
}

var baseCmd = BaseCMD()

// ConnectCommand - builds ssh command to connect to a remote host.
func ConnectCommand(options ...Option) string {
	sb := strings.Builder{}
	sb.WriteString(baseCmd)

	for _, option := range options {
		addOption(&sb, option)
	}

	return sb.String()
}

// LoadConfigCommand - builds ssh command to load config from ssh_config file.
func LoadConfigCommand(option OptionReadConfig) string {
	sb := strings.Builder{}
	sb.WriteString(baseCmd)

	addOption(&sb, option)

	return sb.String()
}
