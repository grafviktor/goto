//go:build windows

package sshconfig

// BaseCMD return OS specific 'ssh' command.
func BaseCMD() string {
	return "cmd /c ssh"
}
