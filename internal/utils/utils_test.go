package utils //nolint:revive,nolintlint // utils is a common name

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"testing"
	"time"

	"github.com/samber/lo"
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
	tmpFile, _ := os.CreateTemp(t.TempDir(), "unit_test_tmp.")
	defer tmpFile.Close()
	err := CreateAppDirIfNotExists(tmpFile.Name())
	require.Error(
		t,
		err,
		"CreateAppDirIfNotExists should return an error when home path exists and it's not a directory",
	)

	err = CreateAppDirIfNotExists(" ")
	require.Error(t, err, "CreateAppDirIfNotExists should return an error when argument is empty")

	tmpDir := t.TempDir()
	err = CreateAppDirIfNotExists(tmpDir)
	require.NoError(t, err, "CreateAppDirIfNotExists should not return an error when app home exists")

	tmpDir = path.Join(t.TempDir(), "test")
	err = CreateAppDirIfNotExists(tmpDir)
	require.NoError(t, err, "CreateAppDirIfNotExists should create app home folder if not exists")
}

func Test_GetAppDir(t *testing.T) {
	userConfigDir, _ := os.UserConfigDir()

	expected := path.Join(userConfigDir, "test")
	got, _ := AppDir("test", "")
	require.Equal(t, expected, got, "Should create a subfolder with a certain name in user config directory")

	expected, _ = filepath.Abs(".")
	got, _ = AppDir("test", ".")
	require.Equal(t, expected, got, "Should ignore application name and use a user-defined folder")

	tmp, _ := os.CreateTemp(t.TempDir(), "unit_test_tmp*")
	defer tmp.Close()
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

	// Test case: custom file path
	tempFile, err := os.CreateTemp(t.TempDir(), "ssh_config_tmp.")
	require.NoError(t, err, "Should create a temporary file for testing")
	defer tempFile.Close() // clean up
	customPath := tempFile.Name()
	got, err = SSHConfigFilePath(customPath)
	require.NoError(t, err, "Should not return any errors because the path is valid")
	require.Equal(t, customPath, got, "Should return custom ssh config file path")

	// Test case: custom file path - unsupported URL
	_, err = SSHConfigFilePath("http://127.0.0.1/ssh_config")
	require.NoError(t, err, "Should not return any errors because that's a valid URL")
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

	require.NoError(t, err)
	// Make sure that 'n' is equal to the data length which we sent to the writer
	require.Equal(t, len(data), n)
	// However we can read the text from writer.Output variable when we need
	require.Equal(t, data, writer.Output)
}

func Test_IsURL(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "Valid HTTPS URL",
			input:    "https://example.com/path",
			expected: true,
		},
		{
			name:     "Valid HTTP URL",
			input:    "http://example.com",
			expected: true,
		},
		{
			name:     "Invalid URL - no protocol",
			input:    "www.example.com/path",
			expected: false,
		},
		{
			name:     "Invalid URL - random string",
			input:    "not a url",
			expected: false,
		},
		{
			name:     "Invalid URL - empty string",
			input:    "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsSupportedURL(tt.input)
			require.Equal(t, tt.expected, result)
		})
	}
}

func Test_ExtractBaseURL(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    string
		expectError bool
	}{
		{
			name:        "HTTPS URL with path and query",
			input:       "https://127.0.0.1:8080/path/to/resource?query=value",
			expected:    "https://127.0.0.1:8080",
			expectError: false,
		},
		{
			name:        "HTTP URL with path",
			input:       "http://127.0.0.1/api/v1/users",
			expected:    "http://127.0.0.1",
			expectError: false,
		},
		{
			name:        "Invalid URL - no protocol",
			input:       "127.0.0.1/path",
			expected:    "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ExtractBaseURL(tt.input)

			if tt.expectError {
				require.Error(t, err)
				require.Equal(t, tt.expected, result)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expected, result)
			}
		})
	}
}

func Test_FetchFromURL(t *testing.T) {
	// using a reduced network timeout as we don't want to wait too long when running unit tests
	networkResponseTimeout = time.Second
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/test_1" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("hello world"))
			return
		}

		if r.URL.Path == "/test_2" {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if r.URL.Path == "/test_3" {
			// Sleep longer than FetchFromURL's context timeout
			time.Sleep(networkResponseTimeout + time.Second)
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("late response"))
			return
		}
	}))
	defer ts.Close()

	tests := []struct {
		name          string
		url           string
		expectedData  string
		expectedError error
	}{
		{
			name:          "won't reach server",
			url:           "www.missing_protocol_url.com",
			expectedData:  "",
			expectedError: errors.New("not a valid URL: www.missing_protocol_url.com"),
		},
		{
			name:          "test_1",
			url:           ts.URL + "/test_1",
			expectedData:  "hello world",
			expectedError: nil,
		},
		{
			name:          "test_2",
			url:           ts.URL + "/test_2",
			expectedData:  "",
			expectedError: errors.New("failed to fetch URL " + ts.URL + "/test_2: status code 500"),
		},
		{
			name:          "test_3",
			url:           ts.URL + "/test_3",
			expectedData:  "late response",
			expectedError: context.DeadlineExceeded,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := FetchFromURL(tt.url)
			switch {
			case tt.name == "test_3":
				require.Error(t, err)
				require.ErrorIs(t, err, context.DeadlineExceeded)
			case tt.expectedError != nil:
				require.Error(t, err)
				require.Equal(t, err.Error(), tt.expectedError.Error())
			default:
				defer resp.Close()
				require.NoError(t, err)
				data, readErr := io.ReadAll(resp)
				require.NoError(t, readErr)
				require.Equal(t, tt.expectedData, string(data))
			}
		})
	}
}
