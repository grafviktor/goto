// Package main contains the application entry point for the GOTO SSH Manager.
//
//nolint:lll,gochecknoglobals // Disable line length check, Ignore burn in these variables.
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/samber/lo"

	"github.com/grafviktor/goto/internal/application"
	"github.com/grafviktor/goto/internal/config"
	"github.com/grafviktor/goto/internal/logger"
	"github.com/grafviktor/goto/internal/storage"
	"github.com/grafviktor/goto/internal/ui"
	"github.com/grafviktor/goto/internal/ui/theme"
	"github.com/grafviktor/goto/internal/utils"
	"github.com/grafviktor/goto/internal/version"
)

var (
	buildVersion string
	buildDate    string
	buildCommit  string
	buildBranch  string
)

const (
	appName             = "goto"
	defaultThemeName    = "default"
	logMsgCloseApp      = "--------= Close application =-------"
	logMsgCloseAppError = "--------= Close application with non-zero code =--------"
	exitCodeError       = 1
	exitCodeSuccess     = 0
)

type loggerInterface interface {
	Debug(format string, args ...any)
	Warn(format string, args ...any)
	Info(format string, args ...any)
	Error(format string, args ...any)
	Close()
}

func main() {
	// Set application version and build details
	version.Set(buildVersion, buildCommit, buildBranch, buildDate)

	// Create application configuration
	config, configUpdateStatus, err := config.New()
	if err != nil {
		log.Fatalf("[MAIN] Error: %v", err)
	}

	// Create prerequisites
	err = checkRequirements(config)
	if err != nil {
		log.Fatalf("[MAIN] Error: %v", err)
	}

	// Create application logger
	fileLogger, err := logger.New(config.AppHome, config.LogLevel)
	if err != nil {
		log.Fatalf("[MAIN] Error: %v", err)
	}

	fileLogger.Info("[MAIN] Start application")
	fileLogger.Debug("[MAIN] Parameters: %+v", os.Args[1:])
	version.LogDetails(fileLogger)
	config.LogDetails(fileLogger)

	if config.ShouldExitAfterConfigChange {
		logCloseAndExit(fileLogger, exitCodeSuccess, fmt.Sprintf("[MAIN] %s", configUpdateStatus))
	}

	// Create state
	appState, err := application.New(context.Background(), config, fileLogger)
	if err != nil {
		logMessage := fmt.Sprintf("[MAIN] Cannot initialize application state: %v", err)
		logCloseAndExit(fileLogger, exitCodeError, logMessage)
	}

	// ---- Main application flow ---- //

	// Init storage
	str, err := storage.Initialize(appState.Context, config, appState.Logger)
	if err != nil {
		logMessage := fmt.Sprintf("[MAIN] Cannot access application storage: %v", err)
		logCloseAndExit(fileLogger, exitCodeError, logMessage)
	}

	// Initialize themes
	fileLogger.Debug("[MAIN] Load application theme")
	themeName := lo.Ternary(utils.StringEmpty(&appState.Theme), defaultThemeName, appState.Theme)
	appTheme := theme.LoadTheme(appState.ApplicationConfig.AppHome, themeName, fileLogger)
	fileLogger.Debug("[MAIN] Use theme: %s", appTheme.Name)

	// Run user interface
	if err = ui.Start(appState.Context, str, appState); err != nil {
		logMessage := fmt.Sprintf("[MAIN] Error: %v", err)
		str.Close()
		logCloseAndExit(fileLogger, exitCodeError, logMessage)
	}

	// Quit signal should be intercepted on the UI level, however it will require
	// additional switch-case block with appropriate checks. Leaving this message here.
	fileLogger.Debug("[MAIN] Receive quit signal")
	fileLogger.Debug("[MAIN] Close storage")
	str.Close()
	fileLogger.Debug("[MAIN] Save application state")
	if err = appState.Persist(); err != nil {
		logMessage := fmt.Sprintf("[MAIN] Can't save application state before closing: %v", err)
		logCloseAndExit(fileLogger, exitCodeError, logMessage)
	}

	logCloseAndExit(fileLogger, exitCodeSuccess, "")
}

func checkRequirements(config *config.Configuration) error {
	var err error

	// Check if "ssh" utility is in application path
	if err = utils.CheckAppInstalled("ssh"); err != nil {
		return fmt.Errorf("ssh utility is not installed or cannot be found in the executable path: %w", err)
	}

	// Set application home folder path
	if config.AppHome, err = utils.AppDir(appName, config.AppHome); err != nil {
		log.Printf("[MAIN] Cannot access application home folder: %v", err)

		err = utils.CreateAppDirIfNotExists(config.AppHome)
		if err != nil {
			return fmt.Errorf("cannot create application home folder: %w", err)
		}
	}

	return nil
}

// logCloseAndExit logs the close message, closes the logger, and exits with the specified code.
func logCloseAndExit(lg loggerInterface, exitCode int, exitReason string) {
	loggingFunc := lo.Ternary(exitCode == exitCodeSuccess, lg.Info, lg.Error)
	closeMsg := lo.Ternary(exitCode == exitCodeSuccess, logMsgCloseApp, logMsgCloseAppError)

	if !utils.StringEmpty(&exitReason) {
		loggingFunc(exitReason)
	}

	loggingFunc("[MAIN] %s", closeMsg)

	lg.Close()
	os.Exit(exitCode)
}
