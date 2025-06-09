package sshconfig

import (
	"os"
	"path/filepath"
	"testing"
)

func Test_isTextFileMime(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a text file
	textFile := filepath.Join(tmpDir, "file.txt")
	if err := os.WriteFile(textFile, []byte("hello world\nthis is a test"), 0644); err != nil {
		t.Fatalf("failed to create text file: %v", err)
	}

	// Create a binary file
	binFile := filepath.Join(tmpDir, "file.bin")
	if err := os.WriteFile(binFile, []byte{0x00, 0x01, 0x02, 0x03, 0x04}, 0644); err != nil {
		t.Fatalf("failed to create binary file: %v", err)
	}

	// Non-existent file
	nonExistent := filepath.Join(tmpDir, "doesnotexist.txt")

	tests := []struct {
		name     string
		filename string
		want     bool
	}{
		{"text file", textFile, true},
		{"binary file", binFile, false},
		{"non-existent file", nonExistent, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isTextFileMime(tt.filename)
			if got != tt.want {
				t.Errorf("isTextFileMime(%q) = %v, want %v", tt.filename, got, tt.want)
			}
		})
	}
}
