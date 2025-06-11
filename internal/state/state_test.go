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
	"gopkg.in/yaml.v2"

	"github.com/grafviktor/goto/internal/application"
)

type MockLogger struct {
	Logs []string
}

func (ml *MockLogger) print(format string, args ...interface{}) {
	logMessage := format
	if len(args) > 0 {
		logMessage = fmt.Sprintf(format, args...)
	}
	ml.Logs = append(ml.Logs, logMessage)
}

func (l *MockLogger) Debug(format string, args ...any) {
	l.print(format, args...)
}

func (l *MockLogger) Info(format string, args ...any) {
	l.print(format, args...)
}

func (l *MockLogger) Warn(format string, args ...any) {
	l.print(format, args...)
}

func (l *MockLogger) Error(format string, args ...any) {
	l.print(format, args...)
}

func (l *MockLogger) Close() {
}

type mockOnce struct{}

func (m *mockOnce) Do(f func()) {
	f()
}

// Test reading app state
func Test_CreateApplicationState(t *testing.T) {
	// Use a mock to avoid sync.Once restrictions in tests
	once = &mockOnce{}

	// Create a mock logger for testing
	mockLogger := MockLogger{}

	appState := Create(context.TODO(), application.Configuration{}, &mockLogger)

	// Ensure that the application state is not nil
	assert.NotNil(t, appState)

	// Ensure that the logger was called during the initialization.
	// The first line always contains "Get application state"
	assert.Contains(t, mockLogger.Logs[0], "Create application state")
}

func Test_GetApplicationState(t *testing.T) {
	// Use a mock to avoid sync.Once restrictions in tests
	once = &mockOnce{}

	// Create a mock logger for testing
	mockLogger := MockLogger{}

	Create(context.TODO(), application.Configuration{}, &mockLogger)
	appState := Get()

	// Ensure that the application state is not nil
	assert.NotNil(t, appState)
}

// Test persisting app state
func Test_PersistApplicationState(t *testing.T) {
	// Set up a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Create a mock logger for testing
	mockLogger := MockLogger{}

	// Call the Get function with the temporary directory and mock logger
	appState := Create(context.TODO(), application.Configuration{}, &mockLogger)
	appState.appStateFilePath = path.Join(tempDir, "state.yaml")

	// Modify the application state
	appState.Selected = 42

	// Persist the modified state to disk
	err = appState.Persist()
	assert.NoError(t, err)

	// Read the persisted state from disk
	persistedState := &Application{}
	fileData, err := os.ReadFile(path.Join(tempDir, stateFile))
	assert.NoError(t, err)

	err = yaml.Unmarshal(fileData, persistedState)
	assert.NoError(t, err)

	// Ensure that the persisted state matches the modified state
	assert.Equal(t, appState.Selected, persistedState.Selected)
}

// Test persisting app state
func Test_PersistApplicationStateError(t *testing.T) {
	t.Skip()
	// Create a mock logger for testing
	mockLogger := MockLogger{}

	// Call the Get function with the temporary directory and mock logger
	appState := Create(context.TODO(), application.Configuration{}, &mockLogger)
	appState.appStateFilePath = "non_exitent.yaml"

	// Modify the application state
	appState.Selected = 42

	// Persist the modified state to disk
	err := appState.Persist()
	assert.Error(t, err)
}

func Test_PrintConfigTo(t *testing.T) {
	appConfig := application.Configuration{
		AppHome:           "/tmp/goto",
		LogLevel:          "debug",
		SSHConfigFilePath: "/tmp/ssh_config",
	}
	app := &Application{
		ApplicationConfig: appConfig,
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
