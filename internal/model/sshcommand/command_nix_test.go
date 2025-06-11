//go:build !windows

package sshcommand

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBaseCMD(t *testing.T) {
	expected := "ssh"
	result := BaseCMD()

	if result != expected {
		t.Errorf("Expected '%s' but got '%s'", expected, result)
	}
}

func TestCopyIDCommand(t *testing.T) {
	// Rewrite user home, otherwise the test will depend on a username who executes the test
	os.Setenv("HOME", "/home/username")

	expected := "ssh-copy-id -p 2222 -i /home/username/.ssh/id_rsa username@localhost"
	actual := CopyIDCommand(
		OptionAddress{Value: "localhost"},
		OptionLoginName{Value: "username"},
		OptionRemotePort{Value: "2222"},
		OptionPrivateKey{Value: "~/.ssh/id_rsa"},
	)

	require.Equal(t, expected, actual)
}
