//go:build windows

package ssh

// BaseCMD return OS specific 'ssh' command.
func BaseCMD() string {
	return "cmd /c ssh"
}

// CopyIDCommand - builds ssh command to copy ssh key to a remote host.
func CopyIDCommand(options ...Option) string {
	sb := strings.Builder{}
	// FIXME: Not compatible with Windows OS.
	// umask 0007 && umask && some-command
	// type %USERPROFILE%\.ssh\id_rsa.pub | ssh user@host "cat >> .ssh/authorized_keys"
	// cmd /c type %USERPROFILE%\.ssh\id_rsa.pub | ssh user@host "cat >> .ssh/authorized_keys && chmod 600 .ssh/authorized_keys"

	baseCmd := "ssh-copy-id"
	sb.WriteString(baseCmd)

	var hostname string
	var username string
	for _, option := range options {
		switch opt := option.(type) {
		case OptionAddress:
			hostname = opt.Value
		case OptionLoginName:
			username = opt.Value
		case OptionPrivateKey:
			if strings.HasPrefix(opt.Value, "~") {
				// Replace "~" with "$HOME" environment variable
				opt.Value = strings.Replace(opt.Value, "~", os.Getenv("HOME"), 1)
			}
			addOption(&sb, opt)
		default:
			addOption(&sb, opt)
		}
	}

	return fmt.Sprintf("%s %s@%s", sb.String(), username, hostname)
}
