//go:build windows

package host

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/grafviktor/goto/internal/constant"
	"github.com/grafviktor/goto/internal/model/sshconfig"
)

// The only difference between Unix and Windows version tests is that Windows
// version prefixes the original command with "cmd /c "
var osCmdPrefix = "cmd /c "

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
				SSHHostConfig: &sshconfig.Config{
					Hostname:     "localhost",
					IdentityFile: "~/.ssh/test",
					Port:         "2222",
					User:         "root",
				},
			},
			expected: `cmd /c type "C:\Users\username\.ssh\test.pub" | ssh root@localhost -p 2222 "cat >> ~/.ssh/authorized_keys && chmod 600 ~/.ssh/authorized_keys && echo Key added. Now try logging into the machine."`,
		},
		{
			name: "Host loaded from ssh config",
			host: Host{
				Title:       "LOCALHOST_ALIAS",
				Address:     "localhost",
				StorageType: constant.HostStorageType.SSHConfig,
			},
			expected: `cmd /c type "C:\Users\username\.ssh\test.pub" | ssh LOCALHOST_ALIAS "cat >> ~/.ssh/authorized_keys && chmod 600 ~/.ssh/authorized_keys && echo Key added. Now try logging into the machine."`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := tt.host.CmdSSHCopyID()
			require.Equal(t, tt.expected, actual)
		})
	}
}
