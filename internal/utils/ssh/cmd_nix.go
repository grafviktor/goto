//go:build !windows

package ssh

func BaseCMD() string {
	return "ssh"
}
