package utils

import (
	"os"
	"path"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLogDirUsesXDGStateHomeOnLinux(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("XDG state directory is a Linux convention")
	}

	stateHome := t.TempDir()
	t.Setenv("XDG_STATE_HOME", stateHome)

	got, err := LogDir("goto", "/config/goto")
	require.NoError(t, err)
	require.Equal(t, path.Join(stateHome, "goto"), got)
	info, err := os.Stat(got)
	require.NoError(t, err)
	require.True(t, info.IsDir())
}

func TestLogDirFallsBackToLocalStateOnLinux(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("XDG state directory is a Linux convention")
	}

	t.Setenv("XDG_STATE_HOME", "")
	home, err := os.UserHomeDir()
	require.NoError(t, err)

	// Override HOME so the fallback does not write into the real user home.
	t.Setenv("HOME", t.TempDir())
	home, err = os.UserHomeDir()
	require.NoError(t, err)

	got, err := LogDir("goto", "/config/goto")
	require.NoError(t, err)
	require.Equal(t, path.Join(home, ".local", "state", "goto"), got)
}

func TestLogDirKeepsAppPathOnOtherPlatforms(t *testing.T) {
	if runtime.GOOS == "linux" {
		t.Skip("non-Linux platforms keep the config directory")
	}

	got, err := LogDir("goto", "/config/goto")
	require.NoError(t, err)
	require.Equal(t, "/config/goto", got)
}
