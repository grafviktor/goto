//go:build !windows

package host

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/grafviktor/goto/internal/model/sshconfig"
)

// The only difference between Unix and Windows version tests is that Windows
// version prefixes the original command with "cmd /c ".
var osCmdPrefix = ""

func TestCmdSSHCopyID(t *testing.T) {
	t.Setenv("HOME", "/home/username")
	host := Host{
		SSHHostConfig: &sshconfig.Config{
			Hostname:     "localhost",
			IdentityFile: "~/.ssh/test",
			Port:         "2222",
			User:         "root",
		},
	}

	actual := host.CmdSSHCopyID()
	require.Equal(t, "ssh-copy-id -p 2222 -i /home/username/.ssh/test root@localhost", actual)
}
