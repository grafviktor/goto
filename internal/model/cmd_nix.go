//go:build !windows

package model

// BaseCMD return OS specific 'ssh' command.
func BaseCMD() string {
	return "ssh"
}
