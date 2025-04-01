//go:build windows

package sshcommand

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBaseCMD(t *testing.T) {
	expected := "cmd /c ssh"
	result := BaseCMD()

	if result != expected {
		t.Errorf("Expected '%s' but got '%s'", expected, result)
	}
}

func TestCopyIDCommand(t *testing.T) {
	// Rewrite user home, otherwise the test will depend on a username who executes the test
	os.Setenv("USERPROFILE", `c:\Users\username`)

	expected := `cmd /c type "c:\Users\username\.ssh\id_rsa.pub" | ssh username@localhost -p 2222 "cat >> ~/.ssh/authorized_keys && chmod 600 ~/.ssh/authorized_keys && echo Key added. Now try logging into the machine."`
	actual := CopyIDCommand(
		OptionAddress{Value: "localhost"},
		OptionLoginName{Value: "username"},
		OptionRemotePort{Value: "2222"},
		OptionPrivateKey{Value: "~/.ssh/id_rsa"},
	)

	require.Equal(t, expected, actual)
}
