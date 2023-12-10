package version

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSetAndGet(t *testing.T) {
	// Test Set function to update build information
	Set("1.0", "abcdef", "develop", "2023-09-01")

	// Check if the values are correctly updated
	if Number() != "1.0" {
		t.Errorf("Expected BuildVersion() to return '1.0', but got '%s'", Number())
	}

	if BuildDate() != "2023-09-01" {
		t.Errorf("Expected BuildDate() to return '2023-09-01', but got '%s'", BuildDate())
	}

	if CommitHash() != "abcdef" {
		t.Errorf("Expected BuildCommit() to return 'abcdef', but got '%s'", CommitHash())
	}

	if BuildBranch() != "develop" {
		t.Errorf("Expected BuildBranch() to return 'develop', but got '%s'", BuildBranch())
	}
}

func TestPrintConsole(t *testing.T) {
	// Capture the output of PrintConsole
	output := captureOutput(func() {
		Print()
	})

	expectedOutput := fmt.Sprintf("Version:    %s\n", Number())
	expectedOutput += fmt.Sprintf("Commit:     %s\n", CommitHash())
	expectedOutput += fmt.Sprintf("Branch:     %s\n", BuildBranch())
	expectedOutput += fmt.Sprintf("Build date: %s\n", BuildDate())

	require.Equal(t, output, expectedOutput)
}

// captureOutput captures the output of a function and returns it as a string
func captureOutput(f func()) string {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)

	return buf.String()
}
