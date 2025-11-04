// Package main contains the application entry point for the GOTO SSH Manager.
//
//nolint:lll,gochecknoglobals // Disable line length check, Ignore burn in these variables.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/caarlos0/env/v10"
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

func main() {
	// Set application version and build details
	version.Set(buildVersion, buildCommit, buildBranch, buildDate)

	// Create the backbone of the application
	appState := createApplicationOrExit()

	// Init storage
	str, err := storage.Initialize(appState.Context, &appState.ApplicationConfig, appState.Logger)
	if err != nil {
		logMessage := fmt.Sprintf("[MAIN] Cannot access application storage: %v", err)
		logCloseAndExit(appState.Logger, exitCodeError, logMessage)
	}

	// Initialize theme system
	appState.Logger.Debug("[MAIN] Loading application theme")
	themeName := lo.Ternary(utils.StringEmpty(&appState.Theme), defaultThemeName, appState.Theme)
	appTheme := theme.LoadTheme(appState.ApplicationConfig.AppHome, themeName, appState.Logger)
	appState.Logger.Debug("[MAIN] Using theme: %s", appTheme.Name)

	// Run user interface
	if err = ui.Start(appState.Context, str, appState); err != nil {
		logMessage := fmt.Sprintf("[MAIN] Error: %v", err)
		str.Close()
		logCloseAndExit(appState.Logger, exitCodeError, logMessage)
	}

	// Quit signal should be intercepted on the UI level, however it will require
	// additional switch-case block with appropriate checks. Leaving this message here.
	appState.Logger.Debug("[MAIN] Receive quit signal")
	appState.Logger.Debug("[MAIN] Close storage")
	str.Close()
	appState.Logger.Debug("[MAIN] Save application state")
	if err = appState.Persist(); err != nil {
		logMessage := fmt.Sprintf("[MAIN] Can't save application state before closing: %v", err)
		logCloseAndExit(appState.Logger, exitCodeError, logMessage)
	}

	logCloseAndExit(appState.Logger, exitCodeSuccess, "")
}

func createApplicationOrExit() *state.Application {
	// Create application configuration
	applicationConfiguration, success := createConfigurationOrExit()

	// Create application logger.
	lg, err := logger.Create(applicationConfiguration.AppHome, applicationConfiguration.LogLevel)
	if err != nil {
		log.Printf("[MAIN] Can't create log file: %v\n", err)
	}

	lg.Debug("[CONFIG] Set application home folder to %q\n", applicationConfiguration.AppHome)
	lg.Debug("[CONFIG] Set application log level to %q\n", applicationConfiguration.LogLevel)
	lg.Debug("[CONFIG] Enabled features: %q\n", applicationConfiguration.EnableFeature)
	lg.Debug("[CONFIG] Disabled features: %q\n", applicationConfiguration.DisableFeature)
	lg.Debug("[CONFIG] Set SSH config path to %q\n", applicationConfiguration.SSHConfigFilePath)

	// Create application state
	applicationState := state.Create(context.Background(), applicationConfiguration, lg)

	// If "-v" parameter provided, display application version configuration and exit
	if applicationConfiguration.DisplayVersionAndExit {
		lg.Debug("[MAIN] Display application version and exit")
		version.Print()
		applicationState.PrintConfig()
		logCloseAndExit(lg, exitCodeSuccess, "")
	}

	// If "-e" parameter provided, display enabled features and exit
	if applicationConfiguration.EnableFeature != "" {
		err = handleFeatureToggle(lg, applicationState, string(applicationConfiguration.EnableFeature), true)
		if err != nil {
			logMessage := fmt.Sprintf("[MAIN] Cannot save application configuration: %v", err)
			logCloseAndExit(lg, exitCodeError, logMessage)
		}

		logCloseAndExit(lg, exitCodeSuccess, "")
	}

	// If "-d" parameter provided, display disabled features and exit
	if applicationConfiguration.DisableFeature != "" {
		err = handleFeatureToggle(lg, applicationState, string(applicationConfiguration.DisableFeature), false)
		if err != nil {
			logMessage := fmt.Sprintf("[MAIN] Cannot save application configuration: %v", err)
			logCloseAndExit(lg, exitCodeError, logMessage)
		}

		logCloseAndExit(lg, exitCodeSuccess, "")
	}

	// If "-set-theme" parameter provided, set the theme and exit
	if applicationConfiguration.SetTheme != "" {
		lg.Debug("[MAIN] Set application theme and exit")
		err = handleSetTheme(lg, applicationState, applicationConfiguration.SetTheme)
		if err != nil {
			logMessage := fmt.Sprintf("[MAIN] Cannot set theme: %v", err)
			logCloseAndExit(lg, exitCodeError, logMessage)
		}

		logCloseAndExit(lg, exitCodeSuccess, "")
	}

	// Log application version
	lg.Info("[MAIN] Start application")
	lg.Info("[MAIN] Version:    %s", version.Number())
	lg.Info("[MAIN] Commit:     %s", version.CommitHash())
	lg.Info("[MAIN] Branch:     %s", version.BuildBranch())
	lg.Info("[MAIN] Build date: %s", version.BuildDate())

	// Check errors at the very end. That allows to check application version and enable/disable
	// features, even if something is not right with the app.
	if !success {
		logCloseAndExit(lg, exitCodeError, "[MAIN] Exit due to a fatal error. Inspect logs for more details.")
	}

	return applicationState
}

func createConfigurationOrExit() (application.Configuration, bool) {
	envConfig, err := parseEnvironmentConfig()
	if err != nil {
		fmt.Printf("Error parsing environment configuration: %+v\n", err)
	}

	// Check if "ssh" utility is in application path
	if err = utils.CheckAppInstalled("ssh"); err != nil {
		log.Fatalf("[MAIN] ssh utility is not installed or cannot be found in the executable path: %v", err)
	}

	cmdConfig := parseCommandLineFlags(envConfig)
	return setupApplicationConfiguration(cmdConfig)
}

type loggerInterface interface {
	Info(format string, args ...any)
	Error(format string, args ...any)
	Debug(format string, args ...any)
	Close()
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

// handleFeatureToggle handles enabling or disabling features.
func handleFeatureToggle(lg loggerInterface, appState *state.Application, featureName string, enable bool) error {
	action := "Disable"
	if enable {
		action = "Enable"
	}

	lg.Info("[MAIN] %s feature %q and exit", action, featureName)
	fmt.Printf("%sd: '%s'\n", action, featureName)

	if enable {
		appState.SSHConfigEnabled = featureName == featureSSHConfig
	} else {
		appState.SSHConfigEnabled = featureName != featureSSHConfig
	}

	return appState.Persist()
}

// parseEnvironmentConfig parses environment configuration.
func parseEnvironmentConfig() (application.Configuration, error) {
	envConfig := application.Configuration{}
	err := env.Parse(&envConfig)
	return envConfig, err
}

// parseCommandLineFlags parses command line flags and returns the configuration.
func parseCommandLineFlags(envConfig application.Configuration) application.Configuration {
	cmdConfig := application.Configuration{}

	// Command line parameters have the highest precedence
	flag.BoolVar(&cmdConfig.DisplayVersionAndExit, "v", false, "Display application details")
	flag.StringVar(&cmdConfig.AppHome, "f", envConfig.AppHome, "Application home folder")
	flag.StringVar(&cmdConfig.LogLevel, "l", envConfig.LogLevel, "Log verbosity level: debug, info")
	flag.StringVar(
		&cmdConfig.SSHConfigFilePath,
		"s",
		envConfig.SSHConfigFilePath,
		"Specifies an alternative per-user SSH configuration file path",
	)
	flag.Var(
		&cmdConfig.EnableFeature,
		"e",
		fmt.Sprintf("Enable feature. Supported values: %s", strings.Join(application.SupportedFeatures, "|")),
	)
	flag.Var(
		&cmdConfig.DisableFeature,
		"d",
		fmt.Sprintf("Disable feature. Supported values: %s", strings.Join(application.SupportedFeatures, "|")),
	)
	flag.StringVar(&cmdConfig.SetTheme, "set-theme", "", "Set application theme")
	flag.Parse()

	return cmdConfig
}

// handleSetTheme handles setting the application theme.
func handleSetTheme(lg loggerInterface, appState *state.Application, themeName string) error {
	// List available themes
	availableThemes, err := theme.ListAvailableThemes(appState.ApplicationConfig.AppHome, lg)
	if err != nil {
		lg.Error("[CONFIG] Cannot list available themes: %v.", err)
		availableThemes = []string{defaultThemeName}
	}

	// Validate theme name
	themeExists := lo.Contains(availableThemes, themeName)
	if !themeExists {
		logMessage := fmt.Sprintf("Theme %q not found. Available themes: %v", themeName, availableThemes)
		lg.Error("[CONFIG] %s", logMessage)
		logCloseAndExit(lg, exitCodeError, logMessage)
	}

	// Set theme in application state
	appState.Theme = themeName
	lg.Debug("[CONFIG] Set theme to %q", themeName)
	fmt.Printf("Theme set to: '%s'\n", themeName)

	return appState.Persist()
}

// setupApplicationConfiguration sets up the application configuration and validates it.
func setupApplicationConfiguration(config application.Configuration) (application.Configuration, bool) {
	success := true
	var err error

	// Set application home folder path
	config.AppHome, err = utils.AppDir(appName, config.AppHome)
	if err != nil {
		log.Printf("[MAIN] Application home folder error: %v", err)
		success = false
	}

	// Create application folder
	if err = utils.CreateAppDirIfNotExists(config.AppHome); err != nil {
		log.Printf("[MAIN] Can't create application home folder: %v", err)
	} else {
		// Even if there was an error, we created the application home folder.
		success = true
	}

	// Set ssh config file path
	config.IsSSHConfigFilePathDefinedByUser = !utils.StringEmpty(&config.SSHConfigFilePath)
	config.SSHConfigFilePath, err = utils.SSHConfigFilePath(config.SSHConfigFilePath)
	if err != nil {
		log.Printf("[MAIN] Can't open SSH config file: %v", err)
		success = false
	}

	return config, success
}
