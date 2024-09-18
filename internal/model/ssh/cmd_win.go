//go:build windows

package ssh

import (
	"fmt"
	"os"
	"strings"
)

// BaseCMD return OS specific 'ssh' command.
func BaseCMD() string {
	return "cmd /c ssh"
}

// CopyIDCommand - builds ssh command to copy ssh key to a remote host.
func CopyIDCommand(options ...Option) string {
	var hostname string
	var username string
	var remotePort string
	var privateKey string

	for _, option := range options {
		switch opt := option.(type) {
		case OptionAddress:
			hostname = opt.Value
		case OptionLoginName:
			username = opt.Value
		case OptionRemotePort:
			remotePort = opt.Value
		case OptionPrivateKey:
			if strings.HasPrefix(opt.Value, "~") {
				// Replace "~" with "$HOME" environment variable
				opt.Value = strings.Replace(opt.Value, "~", os.Getenv("USERPROFILE"), 1)
				opt.Value = strings.Replace(opt.Value, "/", "\\", -1)
			}
			privateKey = opt.Value
		}
	}

	return fmt.Sprintf("cmd /c type \"%s.pub\" | ssh %s@%s -p %s \"cat >> ~/.ssh/authorized_keys && chmod 600 ~/.ssh/authorized_keys && echo Key added. Now try logging into the machine.\"",
		privateKey,
		username,
		hostname,
		remotePort,
	)
}
