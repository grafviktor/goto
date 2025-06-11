// Package logger incapsulates logger functions
package logger

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"strings"
	"testing"
)

type mockOnce struct{}

func (m *mockOnce) Do(f func()) {
	f()
}

func TestLoggerConstructor(t *testing.T) {
	// Use a mock to avoid sync.Once restrictions in tests
	once = &mockOnce{}
	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	// Set up test cases
	testCases := []struct {
		name             string
		appPath          string
		userSetLogLevel  string
		expectedLogLevel LogLevel
		expectError      bool
	}{
		{"DebugLevel", tmpDir, "debug", LevelDebug, false},
		{"DefaultLevel", tmpDir, "info", LevelInfo, false},
		{"InvalidPath", "/nonexistent", "info", LevelInfo, true},
	}

	// Run tests
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			logger, err := Create(tc.appPath, tc.userSetLogLevel)

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected an error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}

				// Check the log level
				if logger.logLevel != tc.expectedLogLevel {
					t.Errorf("Expected log level %v, but got %v", tc.expectedLogLevel, logger.logLevel)
				}

				// Check if the log file is created
				logFilePath := path.Join(tc.appPath, logFileName)
				_, err := os.Stat(logFilePath)
				if os.IsNotExist(err) {
					t.Errorf("Log file not created at %s", logFilePath)
				}

				// Cleanup: Close the log file
				if logger.logFile != nil {
					logger.logFile.Close()
				}
			}
		})
	}
}

func TestLoggerMethods(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	// Create an appLogger instance for testing
	logger, err := Create(tmpDir, "debug")
	if err != nil {
		t.Fatalf("Failed to create appLogger: %v", err)
	}
	defer logger.Close()

	// Set up test cases
	testCases := []struct {
		name           string
		logLevel       LogLevel
		method         func(format string, args ...interface{})
		expectedPrefix string
		expectOutput   bool
	}{
		{"Debug", LevelDebug, logger.Debug, "DEBG", true},
		{"Info", LevelInfo, logger.Info, "INFO", true},
		{"Warn", LevelInfo, logger.Warn, "WARN", true},
		// LogLevel is 'info', but we're printing debug message, thus it should not be printed.
		{"Debug output should be hushed", LevelInfo, logger.Debug, "", false},
		{"Error", LevelInfo, logger.Error, "ERRO", true},
	}

	// Run tests
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			logger.logLevel = tc.logLevel

			// Redirect logger output to a buffer
			var buf bytes.Buffer
			logger.innerLogger.SetOutput(&buf)

			// Call the method
			tc.method("Test message %d", 42)

			if tc.expectOutput {
				// Check if the log was produced as expected
				expectedLog := fmt.Sprintf("[%s] Test message 42", tc.expectedPrefix)
				if output := buf.String(); !strings.Contains(output, expectedLog) {
					t.Errorf("Expected log output:\n%s\nGot:\n%s", expectedLog, output)
				}
			} else {
				// Check if the log was not produced
				if buf.Len() > 0 {
					t.Errorf("Unexpected log output: %s", buf.String())
				}
			}
		})
	}
}
