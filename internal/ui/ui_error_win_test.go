//go:build windows

package ui

import (
	"errors"
	"syscall"
	"testing"

	"golang.org/x/sys/windows"

	"github.com/grafviktor/goto/internal/testutils/mocklogger"
)

func TestHandleUIStartError(t *testing.T) {
	windowsMsgBox = func(_ windows.HWND, _, _ *uint16, _ uint32) (result int32, err error) {
		// Mock MessageBox to prevent actual pop-ups during tests
		return 0, nil
	}

	tests := []struct {
		name    string
		err     error
		wantLog string
	}{
		{
			name:    "Syscall error",
			err:     syscall.Errno(1),
			wantLog: "[UI] Error starting user interface - syscall error: Incorrect function. (error code: 1)",
		},
		{
			name:    "Syscall error - unsupported terminal type",
			err:     syscall.Errno(windows.ERROR_INVALID_PARAMETER),
			wantLog: "[UI] Error starting user interface - unsupported terminal type or terminal is running in legacy mode.",
		},
		{
			name:    "Non-syscall error",
			err:     errors.New("some other error"),
			wantLog: "[UI] Error starting user interface: some other error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := &mocklogger.Logger{}
			handleUIStartError(tt.err, logger)

			if len(logger.Logs) == 0 {
				t.Fatalf("Expected a log entry, but got none")
			}

			lastLog := logger.Logs[len(logger.Logs)-1]
			if lastLog != tt.wantLog {
				t.Errorf("Expected log format '%s', but got '%s'", tt.wantLog, lastLog)
			}
		})
	}
}
