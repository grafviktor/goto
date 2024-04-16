// Package state is in charge of storing and reading application state.
package state

import (
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"

	"github.com/grafviktor/goto/internal/test"
)

// That's a wrapper function for state.Get which is required to overcome sync.Once restrictions
func stateGet(tempDir string, mockLogger *test.MockLogger) *ApplicationState {
	appState := Get(tempDir, mockLogger)

	// We need this hack because state.Get function utilizes `sync.once`. That means, if all unit tests
	// are ran by a single process, instead of the new tmpDir, the old one will be used. In other words
	// the first test will affect all subsequent tests which rely on state.Get function.
	appState.appStateFilePath = path.Join(tempDir, "state.yaml")

	return appState
}

// Test reading app state
func Test_GetApplicationState(t *testing.T) {
	// Set up a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Create a mock logger for testing
	mockLogger := &test.MockLogger{}

	// Call the Get function with the temporary directory and mock logger
	appState := stateGet(tempDir, mockLogger)

	// Ensure that the application state is not nil
	assert.NotNil(t, appState)

	// Ensure that the logger was called during the initialization.
	// The first line always contains "Get application state"
	assert.Contains(t, mockLogger.Logs[0], "Get application state")
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
	mockLogger := &test.MockLogger{}

	// Call the Get function with the temporary directory and mock logger
	appState := stateGet(tempDir, mockLogger)

	// Modify the application state
	appState.Selected = 42

	// Persist the modified state to disk
	err = appState.Persist()
	assert.NoError(t, err)

	// Read the persisted state from disk
	persistedState := &ApplicationState{}
	fileData, err := os.ReadFile(path.Join(tempDir, stateFile))
	assert.NoError(t, err)

	err = yaml.Unmarshal(fileData, persistedState)
	assert.NoError(t, err)

	// Ensure that the persisted state matches the modified state
	assert.Equal(t, appState.Selected, persistedState.Selected)
}
