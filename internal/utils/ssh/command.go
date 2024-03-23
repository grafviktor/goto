// Package ssh - contains functions to construct ssh command for using when connecting to a remote host
package ssh

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

// ConstructCMD - build connect command from main app and its arguments
// cmd - main executable
// options - set of command line options. See Option... public variables.
func ConstructCMD(cmd string, options ...CommandLineOption) string {
	sb := strings.Builder{}
	sb.WriteString(cmd)

	for _, argument := range options {
		addOption(&sb, argument)
	}

	return sb.String()
}

// Config struct contains values loaded from ~/.ssh_config file. Supported values:
//
// IdentityFile string
// User         string
// Port         string
type Config struct {
	// Values which should be extracted from 'ssh -G <hostname>' command:
	// 1. 'identityfile'
	// 2. 'user'
	// 3. 'port'
	// user roman
	// hostname localhost
	// port 22
	// identity file ~/.ssh/id_rsa
	IdentityFile string
	User         string
	Port         string
}

// ParseConfig - parses 'ssh -G <hostname> command' output and returns Config struct.
func ParseConfig(config string) *Config {
	fmt.Println(config)

	return &Config{}
}
