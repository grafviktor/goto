package ssh

import (
	"fmt"
	"strings"
)

type CommandLineOption interface{}

type (
	OptionPrivateKey struct{ Value string }
	OptionRemotePort struct{ Value string }
	OptionLoginName  struct{ Value string }
	OptionAddress    struct{ Value string }
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
