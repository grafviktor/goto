// Package main contains application entry point
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

	"github.com/grafviktor/goto/internal/application"
	"github.com/grafviktor/goto/internal/logger"
	"github.com/grafviktor/goto/internal/state"
	"github.com/grafviktor/goto/internal/storage"
	"github.com/grafviktor/goto/internal/ui"
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
	appName          = "goto"
	featureSSHConfig = "ssh_config"
	logMsgCloseApp   = "--------= Close application =-------"
)

func main() {
	// Set application version and build details
	version.Set(buildVersion, buildCommit, buildBranch, buildDate)

	// Create the backbone of the application
	appState := createApplicationOrExit()

	// Init storage
	strg, fatalErr := storage.Get(appState.Context, appState.ApplicationConfig, appState.Logger)
	if fatalErr != nil {
		appState.Logger.Error("[MAIN] Cannot access application storage: %v\n", fatalErr)
		os.Exit(1)
	}

	// Run user interface
	ui.Start(appState.Context, strg, &appState)

	// Quit signal should be intercepted on the UI level, however it will require an
	// additional switch-case block with an appropriate checks. Leaving this message here.
	appState.Logger.Debug("[MAIN] Receive quit signal")
	appState.Logger.Debug("[MAIN] Save application state")
	fatalErr = appState.Persist()
	if fatalErr != nil {
		appState.Logger.Error("[MAIN] Can't save application state before closing %v", fatalErr)
	}

	appState.Logger.Info("[MAIN] %s", logMsgCloseApp)
	appState.Logger.Close()
}

func createApplicationOrExit() state.Application {
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

	// Create applicationContext state
	applicationState := state.Create(context.Background(), applicationConfiguration, lg)

	// If "-v" parameter provided, display application version configuration and exit
	if applicationConfiguration.DisplayVersionAndExit {
		lg.Debug("[MAIN] Display application version and exit")
		version.Print()
		fmt.Println()
		applicationState.PrintConfig()
		lg.Debug("[MAIN] %s", logMsgCloseApp)
		os.Exit(0)
	}

	// If "-e" parameter provided, display enabled features and exit
	if applicationConfiguration.EnableFeature != "" {
		lg.Info("[MAIN] Enable feature %q and exit", applicationConfiguration.EnableFeature)
		fmt.Printf("Enabled: '%s'\n", applicationConfiguration.EnableFeature)
		applicationState.SSHConfigEnabled = applicationConfiguration.EnableFeature == featureSSHConfig
		err = applicationState.Persist()
		if err != nil {
			lg.Debug("[MAIN] Cannot save application configuration: %v", err)
		}

		lg.Debug("[MAIN] %s", logMsgCloseApp)
		os.Exit(0)
	}

	// If "-d" parameter provided, display disabled features and exit
	if applicationConfiguration.DisableFeature != "" {
		lg.Info("[MAIN] Disable feature %q and exit", applicationConfiguration.DisableFeature)
		fmt.Printf("Disabled: '%s'\n", applicationConfiguration.DisableFeature)
		applicationState.SSHConfigEnabled = !(applicationConfiguration.DisableFeature == featureSSHConfig)
		err = applicationState.Persist()
		if err != nil {
			lg.Debug("[MAIN] Cannot save application configuration: %v", err)
		}

		lg.Debug("[MAIN] %s", logMsgCloseApp)
		os.Exit(0)
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
		lg.Warn("--------= Close application with non-zero code =--------")
		log.Printf("[MAIN] Exit due to a fatal error.")
		os.Exit(1)
	}

	return *applicationState
}

func createConfigurationOrExit() (application.Configuration, bool) {
	var err error
	success := true

	envConfig := application.Configuration{}
	// Parse environment parameters. These parameters have lower precedence than command line flags
	if err = env.Parse(&envConfig); err != nil {
		fmt.Printf("%+v\n", err)
	}

	// Check if "ssh" utility is in application path
	if err = utils.CheckAppInstalled("ssh"); err != nil {
		log.Fatalf("[MAIN] ssh utility is not installed or cannot be found in the executable path: %v", err)
	}

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
	flag.Parse()

	// Set application home folder path
	cmdConfig.AppHome, err = utils.AppDir(appName, cmdConfig.AppHome)
	if err != nil {
		log.Printf("[MAIN] Application home folder: %v\n", err)
		success = false
	}

	// Create application folder
	if err = utils.CreateAppDirIfNotExists(cmdConfig.AppHome); err != nil {
		log.Printf("[MAIN] Can't create application home folder: %v\n", err)
	} else {
		// Even, if there was an error, we created the application home folder.
		success = true
	}

	// Set ssh config file path
	cmdConfig.SSHConfigFilePath, err = utils.SSHConfigFilePath(cmdConfig.SSHConfigFilePath)
	if err != nil {
		log.Printf("[MAIN] Can't open SSH config file: %v\n", err)
		success = false
	}

	return cmdConfig, success
}
