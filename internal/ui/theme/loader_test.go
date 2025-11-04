package theme

import (
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/grafviktor/goto/internal/testutils/mocklogger"
)

func TestSetTheme(t *testing.T) {
	require.Nil(t, currentTheme)
	SetTheme(DefaultTheme())
	require.Equal(t, "default", currentTheme.Name)
}

func TestGetTheme(t *testing.T) {
	theme := GetTheme()
	require.Equal(t, "default", theme.Name)
}

func TestLoadTheme_ThemesFolderNotExists(t *testing.T) {
	tempDir := t.TempDir()
	theme := LoadTheme(tempDir, "nord", &mocklogger.Logger{})
	require.FileExists(t, path.Join(tempDir, "themes", "nord.json"))
	require.FileExists(t, path.Join(tempDir, "themes", "default.json"))
	require.FileExists(t, path.Join(tempDir, "themes", "solarized-dark.json"))
	require.Equal(t, "nord", theme.Name)
}

func TestLoadTheme_ThemesFolderNotExistsAndCannotCreate(t *testing.T) {
	tempDir := t.TempDir()
	// Here we create a file called "themes" to prevent creating a folder with the same name
	_, _ = os.Create(path.Join(tempDir, "themes"))
	theme := LoadTheme(tempDir, "nord", &mocklogger.Logger{})
	require.NoFileExists(t, path.Join(tempDir, "themes", "nord.json"))
	require.NoFileExists(t, path.Join(tempDir, "themes", "default.json"))
	require.NoFileExists(t, path.Join(tempDir, "themes", "solarized-dark.json"))
	require.Equal(t, "default", theme.Name)
}

func TestLoadTheme_UnknownTheme(t *testing.T) {
	tempDir := t.TempDir()
	theme := LoadTheme(tempDir, "no-such-theme", &mocklogger.Logger{})
	require.FileExists(t, path.Join(tempDir, "themes", "nord.json"))
	require.FileExists(t, path.Join(tempDir, "themes", "default.json"))
	require.FileExists(t, path.Join(tempDir, "themes", "solarized-dark.json"))
	require.Equal(t, "default", theme.Name)
}
