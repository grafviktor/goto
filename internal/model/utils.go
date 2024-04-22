package model

import (
	"fmt"
	"strings"
)

// CommandLineOption - parent interface for command line option.
type CommandLineOption interface{}

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

func addOption(sb *strings.Builder, rawParameter CommandLineOption) {
	var option string
	switch p := rawParameter.(type) {
	case OptionPrivateKey:
		option = constructKeyValueOption("-i", p.Value)
	case OptionRemotePort:
		option = constructKeyValueOption("-p", p.Value)
	case OptionLoginName:
		option = constructKeyValueOption("-l", p.Value)
	case OptionReadConfig:
		option = constructKeyValueOption("-G", p.Value)
	case OptionAddress:
		if p.Value != "" {
			option = fmt.Sprintf(" %s", p.Value)
		}
	default:
		return
	}

	sb.WriteString(option)
}

// hostModelToOptionsAdaptor - extract values from model.Host into a set of ssh.CommandLineOption
// host - model.Host to be adapted
// returns []ssh.CommandLineOption.
func hostModelToOptionsAdaptor(host Host) []CommandLineOption {
	return []CommandLineOption{
		OptionPrivateKey{Value: host.IdentityFilePath},
		OptionRemotePort{Value: host.RemotePort},
		OptionLoginName{Value: host.LoginName},
		OptionAddress{Value: host.Address},
	}
}
