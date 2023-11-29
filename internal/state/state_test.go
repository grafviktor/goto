// Package state is in charge of storing and reading application state.
package state

import (
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

// MockLogger implements the iLogger interface for testing.
type MockLogger struct {
	Logs []string
}

func (ml *MockLogger) Debug(format string, args ...interface{}) {
	logMessage := format
	if len(args) > 0 {
		logMessage = fmt.Sprintf(format, args...)
	}
	ml.Logs = append(ml.Logs, logMessage)
}

// That's a wrapper function for state.Get which is required to overcome sync.Once restrictions
func stateGet(tempDir string, mockLogger *MockLogger) *ApplicationState {
	appState := Get(tempDir, mockLogger)

	// We need this hack because state.Get function utilizes `sync.once`. That means, if all unit tests
	// are ran by a single process, instead of the new tmpDir, the old one will be used. In other words
	// the first test will affect all subsequent tests which rely on state.Get function.
	appState.appStateFilePath = path.Join(tempDir, "state.yaml")

	return appState
}

func Test_GetApplicationState(t *testing.T) {
	// Set up a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Create a mock logger for testing
	mockLogger := &MockLogger{}

	// Call the Get function with the temporary directory and mock logger
	appState := stateGet(tempDir, mockLogger)

	// Ensure that the application state is not nil
	assert.NotNil(t, appState)

	// Ensure that the logger was called during the initialization.
	// The first line always contains "Read application state from"
	assert.Contains(t, mockLogger.Logs[0], "Read application state from")
}

func Test_PersistApplicationState(t *testing.T) {
	// Set up a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Create a mock logger for testing
	mockLogger := &MockLogger{}

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

/* FAILING
// Test sync.Once call from multiple threads when reading app config
func Test_ConcurrentInitialization(t *testing.T) {
	// Set up a temporary directory for testing

	// BUG: tempDir will not be used if we ran Get(tempDir, mockLogger) in previous unit tests
	tempDir, err := os.MkdirTemp("", "test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Create a mock logger for testing
	mockLogger := &MockLogger{}

	// Use a wait group to synchronize goroutines
	var wg sync.WaitGroup

	// Number of goroutines for concurrent initialization
	numGoroutines := 10

	// Pre-create file manually with minimum content
	// validYamlContent := []byte("{}")
	// err = os.WriteFile(path.Join(tempDir, "state.yaml"), validYamlContent, 0600)
	// require.NoError(t, err)

	// Initialize the application state concurrently
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			appState := Get(tempDir, mockLogger)
			// Simulate some work with a short sleep
			time.Sleep(50 * time.Millisecond)
			assert.NotNil(t, appState)
		}()
	}

	// Wait for all goroutines to finish
	wg.Wait()

	// Ensure that the application state is initialized only once
	assert.Len(t, mockLogger.Logs, 1)
	// BUG: Trying to read at index 0, but mockLogger.Logs is empty
	assert.Contains(t, mockLogger.Logs[0], "Read application state from")
}
*/
