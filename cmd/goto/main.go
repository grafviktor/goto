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
	"github.com/grafviktor/goto/internal/logger"
	"github.com/grafviktor/goto/internal/state"
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
	featureSSHConfig    = "ssh_config"
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
	// applicationConfiguration, applicationConfigErr := createConfigurationOrExit()
	applicationConfiguration, applicationConfigErr := application.Create()

	err := checkRequirements(applicationConfiguration)
	if err != nil {
		log.Fatalf("[MAIN] %v", err)
	}

	// Create application logger
	fileLogger, err := logger.Create(applicationConfiguration.AppHome, applicationConfiguration.LogLevel)
	if err != nil {
		log.Fatalf("[MAIN] %v", err)
	}
	// Log application version
	version.LogDetails(fileLogger)
	// fileLogger.Info("[MAIN] Start application")
	// fileLogger.Info("[MAIN] Version:    %s", version.Number())
	// fileLogger.Info("[MAIN] Commit:     %s", version.CommitHash())
	// fileLogger.Info("[MAIN] Branch:     %s", version.BuildBranch())
	// fileLogger.Info("[MAIN] Build date: %s", version.BuildDate())

	// fileLogger.Debug("[CONFIG] Set application home folder to %q\n", applicationConfiguration.AppHome)
	// fileLogger.Debug("[CONFIG] Set application log level to %q\n", applicationConfiguration.LogLevel)
	// fileLogger.Debug("[CONFIG] Enabled features: %q\n", applicationConfiguration.EnableFeature)
	// fileLogger.Debug("[CONFIG] Disabled features: %q\n", applicationConfiguration.DisableFeature)
	// fileLogger.Debug("[CONFIG] Set SSH config path to %q\n", applicationConfiguration.SSHConfigFilePath)
	// configurationChanged, err := applicationConfiguration.handleCommandLineParameters(fileLogger)

	applicationConfiguration.LogDetails(fileLogger)

	// Create state
	appState, err := state.Create(context.Background(), applicationConfiguration, fileLogger)
	if err != nil {
		logMessage := fmt.Sprintf("[MAIN] Cannot initialize application state: %v", err)
		logCloseAndExit(fileLogger, exitCodeError, logMessage)
	}
	appState.LogDetails(fileLogger)

	// Check config errors at the very end. That allows to check application version and enable/disable
	// features, even if something is not right with the app.
	if applicationConfigErr != nil {
		logMessage := fmt.Sprintf("[MAIN] Exit due to a fatal error: %v. Inspect logs for more details.", applicationConfigErr)
		logCloseAndExit(fileLogger, exitCodeError, logMessage)
	}

	// If configuration was changed due to command-line parameters, exit the application.
	// if configurationChanged {
	// 	logCloseAndExit(fileLogger, exitCodeSuccess, "")
	// }

	// ---- Main application flow ---- //

	// Init storage
	str, err := storage.Initialize(appState.Context, applicationConfiguration, appState.Logger)
	if err != nil {
		logMessage := fmt.Sprintf("[MAIN] Cannot access application storage: %v", err)
		logCloseAndExit(fileLogger, exitCodeError, logMessage)
	}

	// Initialize themes
	fileLogger.Debug("[MAIN] Loading application theme")
	themeName := lo.Ternary(utils.StringEmpty(&appState.Theme), defaultThemeName, appState.Theme)
	appTheme := theme.LoadTheme(appState.ApplicationConfig.AppHome, themeName, fileLogger)
	fileLogger.Debug("[MAIN] Using theme: %s", appTheme.Name)

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

func checkRequirements(config *application.Configuration) error {
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
func logCloseAndExit(lg loggerInterface, exitCode int, errorExitReason string) {
	if exitCode != exitCodeSuccess {
		fmt.Printf("%s\n", errorExitReason)
		lg.Error("[MAIN] %s", logMsgCloseAppError)
	} else {
		lg.Info("[MAIN] %s", logMsgCloseApp)
	}

	lg.Close()
	os.Exit(exitCode)
}

// func handleCommandLineParameters(
// 	lg loggerInterface,
// 	applicationConfiguration *application.Configuration,
// 	applicationState *state.Application,
// ) (bool, error) {
// 	// If "-v" parameter provided, display application version configuration and exit
// 	if applicationConfiguration.DisplayVersionAndExit {
// 		lg.Debug("[MAIN] Display application version and exit")
// 		version.Print()
// 		applicationState.PrintConfig()
// 		// logCloseAndExit(lg, exitCodeSuccess, "")
// 		return true, nil
// 	}

// 	// If "-e" parameter provided, display enabled features and exit
// 	if applicationConfiguration.EnableFeature != "" {
// 		lg.Debug("[MAIN] Enable feature %q and exit", applicationConfiguration.EnableFeature)
// 		err := application.HandleFeatureToggle(lg, applicationState, string(applicationConfiguration.EnableFeature), true)
// 		if err != nil {
// 			// logMessage := fmt.Sprintf("[MAIN] Cannot save application configuration: %v", err)
// 			return true, fmt.Errorf("cannot save application configuration: %w", err)
// 			// logCloseAndExit(lg, exitCodeError, logMessage)
// 		}

// 		// logCloseAndExit(lg, exitCodeSuccess, "")
// 		return true, nil
// 	}

// 	// If "-d" parameter provided, display disabled features and exit
// 	if applicationConfiguration.DisableFeature != "" {
// 		lg.Debug("[MAIN] Disable feature %q and exit", applicationConfiguration.EnableFeature)
// 		err := application.HandleFeatureToggle(lg, applicationState, string(applicationConfiguration.DisableFeature), false)
// 		if err != nil {
// 			// logMessage := fmt.Sprintf("[MAIN] Cannot save application configuration: %v", err)
// 			// logCloseAndExit(lg, exitCodeError, logMessage)
// 			return true, fmt.Errorf("cannot save application configuration: %w", err)
// 		}

// 		// logCloseAndExit(lg, exitCodeSuccess, "")
// 		return true, nil
// 	}

// 	// If "-set-theme" parameter provided, set the theme and exit
// 	if applicationConfiguration.SetTheme != "" {
// 		lg.Debug("[MAIN] Set application theme and exit")
// 		err := application.HandleSetTheme(lg, applicationState, applicationConfiguration.SetTheme)
// 		if err != nil {
// 			// logMessage := fmt.Sprintf("[MAIN] Cannot set theme: %v", err)
// 			// logCloseAndExit(lg, exitCodeError, logMessage)
// 			return true, fmt.Errorf("cannot set theme: %w", err)
// 		}

// 		return true, nil
// 	}

// 	return false, nil
// }
