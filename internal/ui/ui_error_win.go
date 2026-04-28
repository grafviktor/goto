//go:build windows

package ui

import (
	"errors"
	"fmt"
	"syscall"

	"golang.org/x/sys/windows"
)

func handleUIStartError(err error, logger iLogger) error {
	var errno syscall.Errno
	if errors.As(err, &errno) {
		handleSyscallError(err, errno, logger)
	} else {
		logger.Error("[UI] Error starting user interface: %v", err)
		displayMessageBox("Application Error", fmt.Sprintf("Failed to start user interface: %v", err))
	}

	return err
}

func handleSyscallError(err error, errno syscall.Errno, logger iLogger) {
	if errno == windows.ERROR_INVALID_PARAMETER {
		// See \go\src\internal\syscall\windows\symlink_windows.go: ERROR_INVALID_PARAMETER (code: 87)
		logger.Error("[UI] Error starting user interface - unsupported terminal type or terminal is running in legacy mode.")
		displayMessageBox("Terminal Error", "Unsupported terminal type or\nterminal is running in legacy mode.")
	} else {
		errMsg := fmt.Sprintf("syscall error: %s (error code: %d)", err.Error(), errno)
		logger.Error("[UI] Error starting user interface - %s", errMsg)
		displayMessageBox("System Error", fmt.Sprintf("Terminal interface initialization failed: %s", errMsg))
	}
}

// Store original MessageBox function in a variable to allow mocking in tests
var windowsMsgBox = windows.MessageBox

func displayMessageBox(title, message string) {
	windowsMsgBox(
		0,
		windows.StringToUTF16Ptr(message),
		windows.StringToUTF16Ptr(title),
		windows.MB_OK|windows.MB_ICONERROR,
	)
}
