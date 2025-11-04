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
	cleanup()
}

func TestGetTheme(t *testing.T) {
	require.Nil(t, currentTheme)
	theme := GetTheme()
	require.Equal(t, "default", theme.Name)
	cleanup()
}

func TestLoadTheme_ThemesFolderNotExists(t *testing.T) {
	require.Nil(t, currentTheme)
	tempDir := t.TempDir()
	theme := LoadTheme(tempDir, "nord", &mocklogger.Logger{})
	require.FileExists(t, path.Join(tempDir, "themes", "nord.json"))
	require.FileExists(t, path.Join(tempDir, "themes", "default.json"))
	require.FileExists(t, path.Join(tempDir, "themes", "solarized-dark.json"))
	require.Equal(t, "nord", theme.Name)
	cleanup()
}

func TestLoadTheme_ThemesFolderNotExistsAndCannotCreate(t *testing.T) {
	require.Nil(t, currentTheme)
	tempDir := t.TempDir()
	// Here we create a file called "themes" to prevent creating a folder with the same name
	f, err := os.Create(path.Join(tempDir, "themes"))
	require.NoError(t, err)
	defer f.Close() // Required for Windows, otherwise, it won't be able to delete the test temp folder

	theme := LoadTheme(tempDir, "nord", &mocklogger.Logger{})
	require.NoFileExists(t, path.Join(tempDir, "themes", "nord.json"))
	require.NoFileExists(t, path.Join(tempDir, "themes", "default.json"))
	require.NoFileExists(t, path.Join(tempDir, "themes", "solarized-dark.json"))
	require.Equal(t, "default", theme.Name)
	cleanup()
}

func TestLoadTheme_UnknownTheme(t *testing.T) {
	require.Nil(t, currentTheme)
	tempDir := t.TempDir()
	theme := LoadTheme(tempDir, "no-such-theme", &mocklogger.Logger{})
	require.FileExists(t, path.Join(tempDir, "themes", "nord.json"))
	require.FileExists(t, path.Join(tempDir, "themes", "default.json"))
	require.FileExists(t, path.Join(tempDir, "themes", "solarized-dark.json"))
	require.Equal(t, "default", theme.Name)
	cleanup()
}

/* -------------------------- */

func cleanup() {
	currentTheme = nil
}
