//go:build !windows

package ui

import (
	"errors"
	"fmt"
	"syscall"
	"testing"

	"github.com/grafviktor/goto/internal/testutils/mocklogger"
)

func TestHandleUIStartError(t *testing.T) {
	tests := []struct {
		name    string
		err     error
		wantLog string
	}{
		{
			name: "Syscall error",
			err:  syscall.Errno(1),
			// Had to use fmt.Sprintf because system errors are very specific and can be different across UNIX systems.
			wantLog: fmt.Sprintf("[UI] Error starting user interface - syscall error: %s (error code: %d)",
				syscall.Errno(1).Error(), 1),
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
