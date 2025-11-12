// Package state is in charge of storing and reading application state.
package state

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"

	"github.com/grafviktor/goto/internal/config"
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
func Test_CreateApplicationState(t *testing.T) {
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

func Test_GetApplicationState(t *testing.T) {
	// Use a mock to avoid sync.Once restrictions in tests
	once = &mockOnce{}

	// Create a mock logger for testing
	mockLogger := MockLogger{}

	Initialize(context.TODO(), &config.Configuration{}, &mockLogger)
	underTest := Get()

	// Ensure that the application state is not nil
	assert.NotNil(t, underTest)
}

// Test persisting app state.
func Test_PersistApplicationState(t *testing.T) {
	// Set up a temporary directory for testing
	tempDir := t.TempDir()

	// Create a mock logger for testing
	mockLogger := MockLogger{}

	// Call the Get function with the temporary directory and mock logger
	underTest, _ := Initialize(context.TODO(), &config.Configuration{}, &mockLogger)
	underTest.appStateFilePath = path.Join(tempDir, "state.yaml")

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
	underTest.appStateFilePath = "non_exitent.yaml"

	// Modify the application state
	underTest.Selected = 42

	// Persist the modified state to disk
	err := underTest.Persist()
	assert.Error(t, err)
}

func Test_PrintConfigTo(t *testing.T) {
	appConfig := config.Configuration{
		AppHome:           "/tmp/goto",
		LogLevel:          "debug",
		SSHConfigFilePath: "/tmp/ssh_config",
	}
	app := &State{
		ApplicationConfig: &appConfig,
		SSHConfigEnabled:  true,
	}

	var buf bytes.Buffer
	app.printConfig(&buf)
	output := buf.String()
	assert.Contains(t, output, "App home:           /tmp/goto")
	assert.Contains(t, output, "Log level:          debug")
	assert.Contains(t, output, "SSH config enabled: true")
	assert.Contains(t, output, "SSH config path:    /tmp/ssh_config")
}
