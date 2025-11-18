// Package state is in charge of storing and reading application state.
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

	// Use struct with pointer values to avoid falling back to zero values during saving to YAML.
	type mockState struct {
		Selected         *int    `yaml:"selected"`
		Group            *string `yaml:"group"`
		Theme            *string `yaml:"theme"`
		ScreenLayout     *string `yaml:"screen_layout"`
		SSHConfigEnabled *bool   `yaml:"enable_ssh_config"`
	}

	// Set up a temporary directory for testing
	tempDir := t.TempDir()

	tests := []struct {
		name             string
		stateFileContent string
		expectedState    State
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
			expectedState: State{
				Selected:         999,
				SSHConfigEnabled: true,
				ScreenLayout:     constant.ScreenLayoutCompact,
				Theme:            "dark",
				Group:            "default",
			},
		},
		{
			name: "State file without screen layout",
			stateFileContent: `
selected: 999
enable_ssh_config: true
group: default
theme: dark
`,
			expectedState: State{
				Selected:         999,
				SSHConfigEnabled: true,
				ScreenLayout:     constant.ScreenLayoutDescription,
				Theme:            "dark",
				Group:            "default",
			},
		},
		{
			name: "State file without theme",
			stateFileContent: `
selected: 999
enable_ssh_config: true
group: default
screen_layout: compact
`,
			expectedState: State{
				Selected:         999,
				SSHConfigEnabled: true,
				ScreenLayout:     constant.ScreenLayoutCompact,
				Theme:            "default",
				Group:            "default",
			},
		},
		{
			name: "State file SSH config option disabled",
			stateFileContent: `
selected: 999
enable_ssh_config: false
group: default
theme: dark
screen_layout: compact
`,
			expectedState: State{
				Selected:         999,
				SSHConfigEnabled: false,
				ScreenLayout:     constant.ScreenLayoutCompact,
				Theme:            "dark",
				Group:            "default",
			},
		},
		{
			name: "State file SSH config option not set, should default to enabled",
			stateFileContent: `
selected: 999
group: default
theme: dark
screen_layout: compact
`,
			expectedState: State{
				Selected:         999,
				SSHConfigEnabled: true,
				ScreenLayout:     constant.ScreenLayoutCompact,
				Theme:            "dark",
				Group:            "default",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := os.WriteFile(path.Join(tempDir, stateFile), []byte(tt.stateFileContent), 0644)
			require.NoError(t, err)

			underTest := &State{
				AppHome: tempDir,
				Logger:  &MockLogger{},
			}

			underTest.readFromFile()

			assert.Equal(t, tt.expectedState.Selected, underTest.Selected, "state.Selected value mismatch")
			assert.Equal(t, tt.expectedState.SSHConfigEnabled, underTest.SSHConfigEnabled, "state.SSHConfigEnabled value mismatch")
			assert.Equal(t, tt.expectedState.ScreenLayout, underTest.ScreenLayout, "state.ScreenLayout value mismatch")
			assert.Equal(t, tt.expectedState.Theme, underTest.Theme, "state.Theme value mismatch")
			assert.Equal(t, tt.expectedState.Group, underTest.Group, "state.Group value mismatch")
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
	t.Skip()
	// Create a mock logger for testing
	mockLogger := MockLogger{}

	// Call the Get function with the temporary directory and mock logger
	underTest, _ := Initialize(context.TODO(), &config.Configuration{}, &mockLogger)

	// Modify the application state
	underTest.Selected = 42

	// Persist the modified state to disk
	err := underTest.Persist()
	assert.Error(t, err)
}

func Test_PrintConfig(t *testing.T) {
	state := &State{
		AppHome:           "/tmp/goto",
		LogLevel:          "debug",
		SSHConfigEnabled:  true,
		SSHConfigFilePath: "/tmp/ssh_config",
	}

	actualOutput := captureOutput(state.print)
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
