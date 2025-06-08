package utils

import (
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_stringEmpty(t *testing.T) {
	require.True(t, StringEmpty(lo.ToPtr("")))
	require.True(t, StringEmpty(lo.ToPtr(" ")))
	require.False(t, StringEmpty(lo.ToPtr("test")))
}

func Test_StringAbbreviation(t *testing.T) {
	testMap := map[string]string{
		"":                       "",
		"11":                     "1",
		"3 Rivers, Texas":        "3T",
		"Alexandria, Egypt":      "AE",
		"Arzamas16":              "A1",
		"Atomgrad":               "A",
		"Babylon Iraq":           "BI",
		"Carthage, North Africa": "CA",
		"Sverdlovsk 45":          "S4",
		"NewYork":                "NY",
		"Rio de Janeiro":         "RJ",
		"Thebes_Greece":          "TG",
	}

	for underTest, expected := range testMap {
		actual := StringAbbreviation(underTest)
		require.Equal(t, expected, actual)
	}
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

func Test(t *testing.T) {
	// Test case: SSH config file path
	userConfigDir, _ := os.UserHomeDir()
	expected := path.Join(userConfigDir, ".ssh", "config")
	got, _ := SSHConfigFilePath("")
	require.Equal(t, expected, got, "Should return default ssh config file path")

	// Test case: Path is a directory
	customFolderPath := os.TempDir()
	_, err := SSHConfigFilePath(customFolderPath)
	require.Error(t, err, "SSH config file path is a directory")

	// Test case: custom file path
	tempFile, err := os.CreateTemp(os.TempDir(), "ssh_config*")
	require.NoError(t, err, "Should create a temporary file for testing")
	defer os.Remove(tempFile.Name()) // clean up
	customPath := tempFile.Name()
	got, err = SSHConfigFilePath(customPath)
	require.NoError(t, err, "Should not return any errors because the path is valid")
	require.Equal(t, customPath, got, "Should return custom ssh config file path")
}

func Test_CheckAppInstalled(t *testing.T) {
	tests := []struct {
		name          string
		appName       string
		expectedError bool
	}{
		{
			name:          "Installed App",
			appName:       "find", // Assuming 'find' is always installed
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

func TestBuildConnectSSH(t *testing.T) {
	// Test case: Build SSH sanity check
	cmd := BuildProcessInterceptStdErr("ssh localhost")

	// Check that cmd is created and stdErr is re-defined
	require.NotNil(t, cmd)
	require.Equal(t, os.Stdout, cmd.Stdout)
	require.Equal(t, &ProcessBufferWriter{}, cmd.Stderr)
}

func TestBuildLoadSSHConfig(t *testing.T) {
	// Test case: Load SSH config sanity check
	cmd := BuildProcessInterceptStdAll("localhost")

	// Check that cmd is created and stdErr and stdOut are re-defined
	require.NotNil(t, cmd)
	require.Equal(t, &ProcessBufferWriter{}, cmd.Stderr)
	require.Equal(t, &ProcessBufferWriter{}, cmd.Stdout)
}

func TestSplitArguments(t *testing.T) {
	arguments := `ssh user@127.0.0.1 -o ProxyCommand="/usr/bin/nc -x 127.0.0.1:9689 %h %p" -i /Users/roman/.ssh/id_rsa`
	expected := []string{
		"ssh",
		"user@127.0.0.1",
		"-o",
		"ProxyCommand=/usr/bin/nc -x 127.0.0.1:9689 %h %p",
		"-i",
		"/Users/roman/.ssh/id_rsa",
	}

	actual := splitArguments(arguments)
	require.Equal(t, expected, actual)
}

func Test_ProcessBufferWriter_Write(t *testing.T) {
	// Test the Write method of ProcessBufferWriter
	writer := ProcessBufferWriter{}
	data := []byte("test test test")
	n, err := writer.Write(data)

	assert.NoError(t, err)
	// Make sure that 'n' is equal to the data length which we sent to the writer
	assert.Equal(t, len(data), n)
	// However we can read the text from writer.Output variable when we need
	assert.Equal(t, data, writer.Output)
}
