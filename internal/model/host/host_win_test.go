//go:build windows

package host

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/grafviktor/goto/internal/model/sshconfig"
)

// The only difference between Unix and Windows version tests is that Windows
// version prefixes the original command with "cmd /c "
var windowsCmdPrefix = "cmd /c"

func TestIsUserDefinedSSHCommand(t *testing.T) {
	tests := []struct {
		name     string
		host     Host
		expected string
	}{
		{
			name: "NOT user defined ssh command",
			host: Host{
				Address:          "localhost",
				RemotePort:       "2222",
				LoginName:        "root",
				IdentityFilePath: "/tmp",
			},
			expected: fmt.Sprintf("%s %s", windowsCmdPrefix, "ssh -i /tmp -p 2222 -l root localhost"),
		},
		{
			name: "User defined ssh command",
			host: Host{
				Address: "username@localhost",
			},
			expected: fmt.Sprintf("%s %s", windowsCmdPrefix, "ssh username@localhost"),
		},
		{
			name: "User defined ssh command - other parameters ignored",
			host: Host{
				Address:          "username@localhost",
				RemotePort:       "2222",
				LoginName:        "root",
				IdentityFilePath: "/tmp",
			},
			expected: fmt.Sprintf("%s %s", windowsCmdPrefix, "ssh username@localhost"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := tt.host.CmdSSHConnect()
			require.Equal(t, tt.expected, actual)
		})
	}
}

func TestCmdSSHConfig(t *testing.T) {
	tests := []struct {
		name     string
		host     Host
		expected string
	}{
		{
			name: "NOT user defined ssh command",
			host: Host{
				Address:          "localhost",
				RemotePort:       "2222",
				LoginName:        "root",
				IdentityFilePath: "/tmp",
			},
			expected: fmt.Sprintf("%s %s", windowsCmdPrefix, "ssh -i /tmp -p 2222 -l root -G localhost"),
		},
		{
			name: "User defined ssh command",
			host: Host{
				Address: "username@localhost",
			},
			expected: fmt.Sprintf("%s %s", windowsCmdPrefix, "ssh -G username@localhost"),
		},
		{
			name: "User defined ssh command - other parameters ignored",
			host: Host{
				Address:          "username@localhost",
				RemotePort:       "2222",
				LoginName:        "root",
				IdentityFilePath: "/tmp",
			},
			expected: fmt.Sprintf("%s %s", windowsCmdPrefix, "ssh -G username@localhost"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := tt.host.CmdSSHConfig()
			require.Equal(t, tt.expected, actual)
		})
	}
}

func TestCmdSSHCopyID(t *testing.T) {
	os.Setenv("USERPROFILE", `C:\Users\username`)
	tests := []struct {
		name     string
		host     Host
		expected string
	}{
		{
			name: "NOT user defined ssh command",
			host: Host{
				SSHClientConfig: &sshconfig.Config{
					Hostname:     "localhost",
					IdentityFile: "~/.ssh/test",
					Port:         "2222",
					User:         "root",
				},
			},
			expected: `cmd /c type "C:\Users\username\.ssh\test.pub" | ssh root@localhost -p 2222 "cat >> ~/.ssh/authorized_keys && chmod 600 ~/.ssh/authorized_keys && echo Key added. Now try logging into the machine."`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := tt.host.CmdSSHCopyID()
			require.Equal(t, tt.expected, actual)
		})
	}
}
