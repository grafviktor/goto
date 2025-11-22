package theme

import (
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/grafviktor/goto/internal/testutils/mocklogger"
)

func Test_SetTheme(t *testing.T) {
	require.Nil(t, currentTheme)
	Set(DefaultTheme())
	require.Equal(t, "default", currentTheme.Name)
	cleanup()
}

func Test_GetTheme(t *testing.T) {
	require.Nil(t, currentTheme)
	theme := Get()
	require.Equal(t, "default", theme.Name)
	cleanup()
}

func Test_LoadTheme_ThemesFolderNotExists(t *testing.T) {
	require.Nil(t, currentTheme)
	tempDir := t.TempDir()
	err := Load(tempDir, "nord", &mocklogger.Logger{})
	require.NoError(t, err)
	require.FileExists(t, path.Join(tempDir, "themes", "nord.json"))
	require.FileExists(t, path.Join(tempDir, "themes", "default.json"))
	require.FileExists(t, path.Join(tempDir, "themes", "solarized-dark.json"))
	require.Equal(t, "nord", currentTheme.Name)
	cleanup()
}

func Test_LoadTheme_ThemesFolderNotExistsAndCannotCreate(t *testing.T) {
	require.Nil(t, currentTheme)
	tempDir := t.TempDir()
	// Here we create a file called "themes" to prevent creating a folder with the same name
	f, err := os.Create(path.Join(tempDir, "themes"))
	require.NoError(t, err)
	defer f.Close() // Required for Windows, otherwise, it won't be able to delete the test temp folder

	err = Load(tempDir, "nord", &mocklogger.Logger{})
	require.Error(t, err)
	require.NoFileExists(t, path.Join(tempDir, "themes", "nord.json"))
	require.NoFileExists(t, path.Join(tempDir, "themes", "default.json"))
	require.NoFileExists(t, path.Join(tempDir, "themes", "solarized-dark.json"))
	require.Equal(t, "default", currentTheme.Name)
	cleanup()
}

func Test_LoadTheme_UnknownTheme(t *testing.T) {
	require.Nil(t, currentTheme)
	tempDir := t.TempDir()
	err := Load(tempDir, "no-such-theme", &mocklogger.Logger{})
	require.Error(t, err)
	require.FileExists(t, path.Join(tempDir, "themes", "nord.json"))
	require.FileExists(t, path.Join(tempDir, "themes", "default.json"))
	require.FileExists(t, path.Join(tempDir, "themes", "solarized-dark.json"))
	require.Equal(t, "default", currentTheme.Name)
	cleanup()
}

func cleanup() {
	currentTheme = nil
}

func Test_extractThemeFiles(t *testing.T) {
	themeDir := path.Join(t.TempDir(), "themes")
	logger := &mocklogger.Logger{}
	extractThemeFiles(themeDir, logger)

	require.FileExists(t, path.Join(themeDir, "default.json"))
	require.FileExists(t, path.Join(themeDir, "nord.json"))
	require.FileExists(t, path.Join(themeDir, "solarized-dark.json"))
}

func Test_ListInstalled(t *testing.T) {
	// Test case - themes folder does not exist, but is created by the function
	tempDir := t.TempDir()
	logger := &mocklogger.Logger{}
	themes := ListInstalled(tempDir, logger)

	require.Contains(t, themes, "default")
	require.Contains(t, themes, "nord")
	require.Contains(t, themes, "solarized-dark")

	// Test case - cannot access theme folder
	tempdir2 := path.Join(t.TempDir(), "themes")
	os.WriteFile(tempdir2, []byte("not a folder"), 0o644)
	themes = ListInstalled(tempdir2, logger)

	require.Len(t, themes, 1)
	require.Equal(t, "default", themes[0])
}
