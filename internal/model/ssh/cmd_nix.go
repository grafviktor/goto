//go:build !windows

package ssh

// BaseCMD return OS specific 'ssh' command.
func BaseCMD() string {
	return "ssh"
}
