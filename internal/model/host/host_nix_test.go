//go:build !windows

package host

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/grafviktor/goto/internal/constant"
	"github.com/grafviktor/goto/internal/model/sshconfig"
)

// The only difference between Unix and Windows version tests is that Windows
// version prefixes the original command with "cmd /c "
var osCmdPrefix = ""

func TestCmdSSHCopyID(t *testing.T) {
	tests := []struct {
		name     string
		host     Host
		expected string
	}{
		{
			name: "NOT user defined ssh command",
			host: Host{
				SSHHostConfig: &sshconfig.Config{
					Hostname:     "localhost",
					IdentityFile: "/tmp",
					Port:         "2222",
					User:         "root",
				},
			},
			expected: "ssh-copy-id -p 2222 -i /tmp root@localhost",
		},
		{
			name: "Host loaded from ssh config",
			host: Host{
				Title:       "LOCALHOST_ALIAS",
				Address:     "localhost",
				StorageType: constant.HostStorageType.SSHConfig,
			},
			expected: "ssh-copy-id LOCALHOST_ALIAS",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := tt.host.CmdSSHCopyID()
			require.Equal(t, tt.expected, actual)
		})
	}
}
