//go:build windows

package ssh

func SSHCmd() []string {
	return []string{"cmd", "/c", "ssh"}
}
