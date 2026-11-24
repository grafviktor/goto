package utils

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSplitArgumentsMultibyte(t *testing.T) {
	require.Equal(t, []string{"服务器"}, splitArguments("服务器"))
	require.Equal(t, []string{"ssh", "café"}, splitArguments("ssh café"))
}

func TestBuildProcessMultibyteDoesNotPanic(t *testing.T) {
	cmd := BuildProcess("服务器")
	require.NotNil(t, cmd)
	require.Equal(t, "服务器", cmd.Path)
}
