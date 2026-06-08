//go:build windows

package host

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/grafviktor/goto/internal/model/sshconfig"
)

// The only difference between Unix and Windows version tests is that Windows
// version prefixes the original command with "cmd /c "
var osCmdPrefix = "cmd /c "

func TestCmdSSHCopyID(t *testing.T) {
	t.Setenv("USERPROFILE", `C:\Users\username`)
	host := Host{
		SSHHostConfig: &sshconfig.Config{
			Hostname:     "localhost",
			IdentityFile: "~/.ssh/test",
			Port:         "2222",
			User:         "root",
		},
	}

	actual := host.CmdSSHCopyID()
	expected := `cmd /c type "C:\Users\username\.ssh\test.pub" | ssh root@localhost -p 2222 "cat >> ~/.ssh/authorized_keys && chmod 600 ~/.ssh/authorized_keys && echo Key added. Now try logging into the machine."`
	require.Equal(t, expected, actual)
}
