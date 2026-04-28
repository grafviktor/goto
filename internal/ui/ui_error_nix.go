//go:build !windows

package ui

import (
	"errors"
	"fmt"
	"syscall"
)

func handleUIStartError(err error, logger iLogger) error {
	var errno syscall.Errno
	if errors.As(err, &errno) {
		errMsg := fmt.Sprintf("syscall error: %s (error code: %d)", err.Error(), errno)
		logger.Error("[UI] Error starting user interface - %s", errMsg)
	} else {
		logger.Error("[UI] Error starting user interface: %v", err)
	}

	return err
}
