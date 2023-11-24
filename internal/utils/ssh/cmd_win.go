//go:build windows

package ssh

func BaseCMD() string {
	return "cmd /c ssh"
}
