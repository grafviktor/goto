// Package state is in charge of storing and reading application state.
//
//nolint:golines // Don't care about line length in unit tests.
package state

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"

	"github.com/grafviktor/goto/internal/config"
	"github.com/grafviktor/goto/internal/constant"
)

type MockLogger struct {
	Logs []string
}

func (ml *MockLogger) printf(format string, args ...interface{}) {
	logMessage := format
	if len(args) > 0 {
		logMessage = fmt.Sprintf(format, args...)
	}
	ml.Logs = append(ml.Logs, logMessage)
}

func (ml *MockLogger) Debug(format string, args ...any) {
	ml.printf(format, args...)
}

func (ml *MockLogger) Info(format string, args ...any) {
	ml.printf(format, args...)
}

func (ml *MockLogger) Warn(format string, args ...any) {
	ml.printf(format, args...)
}

func (ml *MockLogger) Error(format string, args ...any) {
	ml.printf(format, args...)
}

func (ml *MockLogger) Close() {
}

type mockOnce struct{}

func (m *mockOnce) Do(f func()) {
	f()
}

// Test reading app state.
func Test_Initialize(t *testing.T) {
	// Use a mock to avoid sync.Once restrictions in tests
	once = &mockOnce{}

	// Create a mock logger for testing
	mockLogger := MockLogger{}

	underTest, _ := Initialize(context.TODO(), &config.Configuration{}, &mockLogger)

	// Ensure that the application state is not nil
	assert.NotNil(t, underTest)

	// Ensure that the logger was called during the initialization.
	// The first line always contains "Get application state"
	assert.Contains(t, mockLogger.Logs[0], "Create application state")
}

func Test_Get(t *testing.T) {
	// Use a mock to avoid sync.Once restrictions in tests
	once = &mockOnce{}

	// Create a mock logger for testing
	mockLogger := MockLogger{}

	Initialize(context.TODO(), &config.Configuration{}, &mockLogger)
	underTest := Get()

	// Ensure that the application state is not nil
	assert.NotNil(t, underTest)
}

func Test_readFromFile(t *testing.T) {
	// Use a mock to avoid sync.Once restrictions in tests
	once = &mockOnce{}

	// Set up a temporary directory for testing
	tempDir := t.TempDir()

	tests := []struct {
		name             string
		stateFileContent string
		expected         State
	}{
		{
			name: "State file with all fields set",
			stateFileContent: `
selected: 999
enable_ssh_config: true
group: default
theme: dark
screen_layout: compact
`,
			expected: State{
				Selected:         999,
				SSHConfigEnabled: true,
				ScreenLayout:     constant.ScreenLayoutCompact,
				Theme:            "dark",
				Group:            "default",
			},
		}, {
			name: "State file without screen layout",
			stateFileContent: `
selected: 999
enable_ssh_config: true
group: default
theme: dark
`,
			expected: State{
				Selected:         999,
				SSHConfigEnabled: true,
				ScreenLayout:     constant.ScreenLayoutDescription,
				Theme:            "dark",
				Group:            "default",
			},
		}, {
			name: "State file without theme",
			stateFileContent: `
selected: 999
enable_ssh_config: true
group: default
screen_layout: compact
`,
			expected: State{
				Selected:         999,
				SSHConfigEnabled: true,
				ScreenLayout:     constant.ScreenLayoutCompact,
				Theme:            "default",
				Group:            "default",
			},
		}, {
			name: "State file SSH config option disabled",
			stateFileContent: `
selected: 999
enable_ssh_config: false
group: default
theme: dark
screen_layout: compact
`,
			expected: State{
				Selected:         999,
				SSHConfigEnabled: false,
				ScreenLayout:     constant.ScreenLayoutCompact,
				Theme:            "dark",
				Group:            "default",
			},
		}, {
			name: "State file SSH config option not set, should default to enabled",
			stateFileContent: `
selected: 999
group: default
theme: dark
screen_layout: compact
`,
			expected: State{
				Selected:         999,
				SSHConfigEnabled: true,
				ScreenLayout:     constant.ScreenLayoutCompact,
				Theme:            "dark",
				Group:            "default",
			},
		}, {
			name: "Valid SSH config path should be picked up by state",
			stateFileContent: `
selected: 999
enable_ssh_config: true
group: default
theme: dark
screen_layout: compact
ssh_config_path: /tmp/some_path
`,
			expected: State{
				Selected:                   999,
				SSHConfigEnabled:           true,
				ScreenLayout:               constant.ScreenLayoutCompact,
				Theme:                      "dark",
				Group:                      "default",
				SSHConfigPath:              "/tmp/some_path",
				SetSSHConfigPath:           "/tmp/some_path",
				IsUserDefinedSSHConfigPath: true,
			},
		}, {
			name: "Valid remote SSH config path should be picked up by state",
			stateFileContent: `
selected: 999
enable_ssh_config: true
group: default
theme: dark
screen_layout: compact
ssh_config_path: http://example.com/ssh_config
`,
			expected: State{
				Selected:                   999,
				SSHConfigEnabled:           true,
				ScreenLayout:               constant.ScreenLayoutCompact,
				Theme:                      "dark",
				Group:                      "default",
				SSHConfigPath:              "http://example.com/ssh_config",
				SetSSHConfigPath:           "http://example.com/ssh_config",
				IsUserDefinedSSHConfigPath: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := os.WriteFile(path.Join(tempDir, stateFile), []byte(tt.stateFileContent), 0o644)
			require.NoError(t, err)

			test := &State{
				AppHome: tempDir,
				Logger:  &MockLogger{},
			}

			test.readFromFile()

			assert.Equal(t, tt.expected.Theme, test.Theme, "state.Theme value mismatch")
			assert.Equal(t, tt.expected.Group, test.Group, "state.Group value mismatch")
			assert.Equal(t, tt.expected.Selected, test.Selected, "state.Selected value mismatch")
			assert.Equal(t, tt.expected.ScreenLayout, test.ScreenLayout, "state.ScreenLayout value mismatch")
			assert.Equal(t, tt.expected.SSHConfigPath, test.SSHConfigPath, "state.SSHConfigPath value mismatch")
			assert.Equal(t, tt.expected.SSHConfigEnabled, test.SSHConfigEnabled, "state.SSHConfigEnabled value mismatch")
			assert.Equal(t, tt.expected.SetSSHConfigPath, test.SetSSHConfigPath, "state.SetSSHConfigPath value mismatch")
			assert.Equal(t, tt.expected.IsUserDefinedSSHConfigPath, test.IsUserDefinedSSHConfigPath, "state.IsUserDefinedSSHConfigPath value mismatch")
		})
	}
}

func Test_applyConfig(t *testing.T) {
	tests := []struct {
		name     string
		testCfg  config.Configuration
		expected State
		wantErr  bool
	}{
		{
			name:    "Empty configuration",
			testCfg: config.Configuration{},
			expected: State{
				AppMode:  constant.AppModeType.StartUI,
				LogLevel: constant.LogLevelType.INFO,
			},
			wantErr: false,
		}, {
			name: "Overwrite LogLevel and AppMode",
			testCfg: config.Configuration{
				AppMode:  constant.AppModeType.DisplayInfo,
				LogLevel: constant.LogLevelType.DEBUG,
			},
			expected: State{
				AppMode:  constant.AppModeType.DisplayInfo,
				LogLevel: constant.LogLevelType.DEBUG,
			},
			wantErr: false,
		}, {
			name:    "Supported feature enabled",
			testCfg: config.Configuration{EnableFeature: "ssh_config"},
			expected: State{
				AppMode:          constant.AppModeType.StartUI,
				LogLevel:         constant.LogLevelType.INFO,
				SSHConfigEnabled: true,
			},
			wantErr: false,
		}, {
			name:    "Supported feature disabled",
			testCfg: config.Configuration{DisableFeature: "ssh_config"},
			expected: State{
				AppMode:          constant.AppModeType.StartUI,
				LogLevel:         constant.LogLevelType.INFO,
				SSHConfigEnabled: false,
			},
			wantErr: false,
		}, {
			name:     "Unsupported feature disabled",
			testCfg:  config.Configuration{DisableFeature: "super_feature"},
			expected: State{},
			wantErr:  true,
		}, {
			name:     "Unsupported feature enabled",
			testCfg:  config.Configuration{EnableFeature: "super_feature"},
			expected: State{},
			wantErr:  true,
		}, {
			name:    "Set SSH config path for current session with '-s' parameter",
			testCfg: config.Configuration{SSHConfigPath: "~/.ssh/custom_config"},
			expected: State{
				AppMode:                    constant.AppModeType.StartUI,
				LogLevel:                   constant.LogLevelType.INFO,
				SSHConfigPath:              "~/.ssh/custom_config",
				IsUserDefinedSSHConfigPath: true,
			},
			wantErr: false,
		}, {
			name:    "Persist SSH config path with '--set-ssh-config-path' parameter",
			testCfg: config.Configuration{SetSSHConfigPath: "~/.ssh/custom_config"},
			expected: State{
				AppMode:          constant.AppModeType.StartUI,
				LogLevel:         constant.LogLevelType.INFO,
				SSHConfigPath:    "",
				SetSSHConfigPath: "~/.ssh/custom_config",
			},
			wantErr: false,
		}, {
			name:    "Persist valid theme '--set-theme' parameter",
			testCfg: config.Configuration{SetTheme: "nord"},
			expected: State{
				AppMode:  constant.AppModeType.StartUI,
				LogLevel: constant.LogLevelType.INFO,
				Theme:    "nord",
			},
			wantErr: false,
		}, {
			name:     "Persist invalid theme '--set-theme' parameter",
			testCfg:  config.Configuration{SetTheme: "no_such_theme"},
			expected: State{},
			wantErr:  true,
		},
	}

	// Use a mock to avoid sync.Once restrictions in tests
	once = &mockOnce{}
	tmpHome := t.TempDir()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.testCfg.AppHome = tmpHome // If remove, it'll extract themes into '.../internal/state/' folder
			actual := &State{Logger: &MockLogger{}}
			err := actual.applyConfig(&tt.testCfg)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected.AppMode, actual.AppMode, "AppMode mismatch")
				assert.Equal(t, tt.expected.LogLevel, actual.LogLevel, "LogLevel mismatch")
				assert.Equal(t, tt.expected.Theme, actual.Theme, "Theme mismatch")
				assert.Equal(t, tt.expected.SSHConfigPath, actual.SSHConfigPath, "SSHConfigPath mismatch")
				assert.Equal(t, tt.expected.SetSSHConfigPath, actual.SetSSHConfigPath, "SetSSHConfigPath mismatch")
				assert.Equal(t, tt.expected.SSHConfigEnabled, actual.SSHConfigEnabled, "SSHConfigEnabled mismatch")
				assert.Equal(t, tt.expected.IsUserDefinedSSHConfigPath, actual.IsUserDefinedSSHConfigPath, "IsUserDefinedSSHConfigPath mismatch")
			}
		})
	}
}

// Test persisting app state.
func Test_PersistApplicationState(t *testing.T) {
	// Set up a temporary directory for testing
	tempDir := t.TempDir()

	// Call the Initialize function with the temporary directory and mock logger
	underTest, _ := Initialize(
		context.TODO(),
		&config.Configuration{AppHome: tempDir},
		&MockLogger{},
	)

	// Modify the application state
	underTest.Selected = 42

	// Persist the modified state to disk
	err := underTest.Persist()
	require.NoError(t, err)

	// Read the persisted state from disk
	persistedState := &State{}
	fileData, err := os.ReadFile(path.Join(tempDir, stateFile))
	require.NoError(t, err)

	err = yaml.Unmarshal(fileData, persistedState)
	require.NoError(t, err)

	// Ensure that the persisted state matches the modified state
	require.Equal(t, underTest.Selected, persistedState.Selected)
}

// Test persisting app state.
func Test_PersistApplicationStateError(t *testing.T) {
	// Create state file with read-only permissions
	appHome := t.TempDir()
	os.WriteFile(path.Join(appHome, "state.yaml"), []byte{}, 0o444)

	// Create a mock logger for testing
	mockLogger := MockLogger{}

	// Call the Get function with the temporary directory and mock logger
	underTest, _ := Initialize(context.TODO(), &config.Configuration{AppHome: appHome}, &mockLogger)

	// Modify the application state
	underTest.Selected = 42

	// Persist the modified state to disk
	err := underTest.Persist()
	assert.Error(t, err)
}

func Test_PrintConfig(t *testing.T) {
	state := &State{
		AppHome:          "/tmp/goto",
		LogLevel:         "debug",
		SSHConfigEnabled: true,
		SSHConfigPath:    "/tmp/ssh_config",
	}

	actualOutput := captureOutput(state.PrintConfig)
	assert.Contains(t, actualOutput, "App home:          /tmp/goto")
	assert.Contains(t, actualOutput, "Log level:         debug")
	assert.Contains(t, actualOutput, "SSH config status: enabled")
	assert.Contains(t, actualOutput, "SSH config path:   /tmp/ssh_config")
}

// captureOutput captures the output of a function and returns it as a string.
func captureOutput(f func()) string {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w //nolint:reassign // For testing purposes

	f()

	w.Close()
	os.Stdout = oldStdout //nolint:reassign // For testing purposes

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)

	return buf.String()
}

func Test_LogDetails(t *testing.T) {
	logger := MockLogger{}
	state := &State{
		AppHome:          "/tmp/goto",
		LogLevel:         "debug",
		SSHConfigEnabled: true,
		SSHConfigPath:    "/tmp/ssh_config",
		Logger:           &logger,
	}

	state.LogDetails()

	assert.Contains(t, logger.Logs[0], `Application home folder: "/tmp/goto"`)
	assert.Contains(t, logger.Logs[1], `Application log level:   "debug"`)
	assert.Contains(t, logger.Logs[2], `SSH config status:       "enabled"`)
	assert.Contains(t, logger.Logs[3], `SSH config path:         "/tmp/ssh_config"`)
}
