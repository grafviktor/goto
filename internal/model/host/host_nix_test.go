//go:build !windows

package host

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/grafviktor/goto/internal/model/ssh"
)

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
			expected: "ssh -i /tmp -p 2222 -l root localhost",
		},
		{
			name: "User defined ssh command",
			host: Host{
				Address: "username@localhost",
			},
			expected: "ssh username@localhost",
		},
		{
			name: "User defined ssh command - other parameters ignored",
			host: Host{
				Address:          "username@localhost",
				RemotePort:       "2222",
				LoginName:        "root",
				IdentityFilePath: "/tmp",
			},
			expected: "ssh username@localhost",
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
			expected: "ssh -i /tmp -p 2222 -l root -G localhost",
		},
		{
			name: "User defined ssh command",
			host: Host{
				Address: "username@localhost",
			},
			expected: "ssh -G username@localhost",
		},
		{
			name: "User defined ssh command - other parameters ignored",
			host: Host{
				Address:          "username@localhost",
				RemotePort:       "2222",
				LoginName:        "root",
				IdentityFilePath: "/tmp",
			},
			expected: "ssh -G username@localhost",
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
	tests := []struct {
		name     string
		host     Host
		expected string
	}{
		{
			name: "NOT user defined ssh command",
			host: Host{
				SSHClientConfig: &ssh.Config{
					Hostname:     "localhost",
					IdentityFile: "/tmp",
					Port:         "2222",
					User:         "root",
				},
			},
			expected: "ssh-copy-id -p 2222 -i /tmp root@localhost",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := tt.host.CmdSSHCopyID()
			require.Equal(t, tt.expected, actual)
		})
	}
}
