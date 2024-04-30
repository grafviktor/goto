package utils

import (
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_stringEmpty(t *testing.T) {
	require.True(t, StringEmpty(""))
	require.True(t, StringEmpty(" "))
	require.False(t, StringEmpty("test"))
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

func Test_CheckAppInstalled(t *testing.T) {
	tests := []struct {
		name          string
		appName       string
		expectedError bool
	}{
		{
			name:          "Installed App",
			appName:       "echo", // Assuming 'echo' is always installed
			expectedError: false,
		},
		{
			name:          "Uninstalled App",
			appName:       "nonexistentapp",
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckAppInstalled(tt.appName)

			if tt.expectedError && err == nil {
				t.Errorf("Expected an error, but got nil")
			}

			if !tt.expectedError && err != nil {
				t.Errorf("Expected no error, but got: %v", err)
			}
		})
	}
}

func TestBuildProcess(t *testing.T) {
	tests := []struct {
		name        string
		cmd         string
		expectedCmd *exec.Cmd
	}{
		{
			name:        "Simple Command",
			cmd:         "cd",
			expectedCmd: exec.Command("cd"),
		},
		{
			name:        "Command with Arguments",
			cmd:         "echo hello",
			expectedCmd: exec.Command("echo", "hello"),
		},
		{
			name:        "Empty Command",
			cmd:         "",
			expectedCmd: nil, // Expecting nil as there is no valid command
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildProcess(tt.cmd)

			switch {
			case tt.expectedCmd == nil && result != nil:
				t.Errorf("Expected nil, but got %+v", result)

			case tt.expectedCmd != nil && result == nil:
				t.Errorf("Expected %+v, but got nil", tt.expectedCmd)
			case tt.expectedCmd != nil && result != nil:
				// Compare relevant fields of the Cmd struct
				if tt.expectedCmd.Path != result.Path {
					t.Errorf("Expected Path %s, but got %s", tt.expectedCmd.Path, result.Path)
				}
				if len(tt.expectedCmd.Args) != len(result.Args) {
					t.Errorf("Expected %d arguments, but got %d", len(tt.expectedCmd.Args), len(result.Args))
				}
				for i := range tt.expectedCmd.Args {
					if tt.expectedCmd.Args[i] != result.Args[i] {
						t.Errorf("Expected argument %s, but got %s", tt.expectedCmd.Args[i], result.Args[i])
					}
				}
			}
		})
	}
}

// func TestBuildConnectSSH(t *testing.T) {
// 	// Test case: Build SSH sanity check
// 	m := model.Host{}
// 	cmd := BuildConnectSSH(m.CmdSSHConnect())
//
// 	// Check that cmd is created and stdErr is re-defined
// 	require.NotNil(t, cmd)
// 	require.Equal(t, os.Stdout, cmd.Stdout)
// 	require.Equal(t, &ProcessBufferWriter{}, cmd.Stderr)
// }
//
// func TestBuildLoadSSHConfig(t *testing.T) {
// 	// Test case: Load SSH config sanity check
// 	cmd := BuildLoadSSHConfig("Mock Host")
//
// 	// Check that cmd is created and stdErr and stdOut are re-defined
// 	require.NotNil(t, cmd)
// 	require.Equal(t, &ProcessBufferWriter{}, cmd.Stderr)
// 	require.Equal(t, &ProcessBufferWriter{}, cmd.Stdout)
// }
