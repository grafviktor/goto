package utils

import (
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_stringEmpty(t *testing.T) {
	require.True(t, stringEmpty(""))
	require.True(t, stringEmpty(" "))
	require.False(t, stringEmpty("test"))
}

func Test_CreateAppDirIfNotExists(t *testing.T) {
	tmpFile, _ := os.CreateTemp("", "unit_test_tmp*")
	defer os.RemoveAll(tmpFile.Name()) // clean up
	err := CreateAppDirIfNotExists(tmpFile.Name())
	require.Error(t, err, "CreateAppDirIfNotExists should return an error when home path exists and it's not a directory")

	err = CreateAppDirIfNotExists(" ")
	require.Error(t, err, "CreateAppDirIfNotExists should return an error when argument is empty")

	tmpDir, _ := os.MkdirTemp("", "unit_test_tmp*")
	defer os.RemoveAll(tmpDir) // clean up
	err = CreateAppDirIfNotExists(tmpDir)
	require.NoError(t, err, "CreateAppDirIfNotExists should not return an error when app home exists")

	tmpDir = path.Join(os.TempDir(), "test")
	defer os.RemoveAll(tmpDir) // clean up
	err = CreateAppDirIfNotExists(tmpDir)
	require.NoError(t, err, "CreateAppDirIfNotExists should create app home folder if not exists")
}

func Test_GetAppDir(t *testing.T) {
	userConfigDir, _ := os.UserConfigDir()

	expected := path.Join(userConfigDir, "test")
	got, _ := AppDir("test", "")
	require.Equal(t, got, expected, "Should create a subfolder with a certain name in user config directory")

	expected, _ = filepath.Abs(".")
	got, _ = AppDir("test", ".")
	require.Equal(t, got, expected, "Should ignore application name and use a user-defined folder")

	tmp, _ := os.CreateTemp("", "unit_test_tmp*")
	defer os.RemoveAll(tmp.Name())
	_, err := AppDir("", tmp.Name())
	require.Error(t, err, "Should not accept file as a user dir")

	_, err = AppDir("", "")
	require.Error(t, err, "App home folder should not be empty 1")

	_, err = AppDir(" ", "")
	require.Error(t, err, "App home folder should not be empty 2")
}

func Test_GetCurrentOSUser(t *testing.T) {
	username := CurrentOSUsername()
	require.NotEmpty(t, username, "GetCurrentOSUser should return a non-empty string")
}
