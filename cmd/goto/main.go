// Package main contains application entry point
//
//nolint:lll // disable line length check
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
		success = false
	}

	lg.Debug("[CONFIG] Set application home folder to '%s'\n", applicationConfiguration.AppHome)
	lg.Debug("[CONFIG] Set application log level to '%s'\n", applicationConfiguration.LogLevel)
	lg.Debug("[CONFIG] Enable feature '%s'\n", applicationConfiguration.EnableFeature)
	lg.Debug("[CONFIG] Disable feature '%s'\n", applicationConfiguration.DisableFeature)
	lg.Debug("[CONFIG] Set SSH config path to '%s'\n", applicationConfiguration.SSHConfigFilePath)

	// Create applicationContext state
	applicationState := state.Create(context.Background(), applicationConfiguration, lg)

	// If "-v" parameter provided, display application version configuration and exit
	if applicationConfiguration.DisplayVersionAndExit {
		lg.Debug("[MAIN] Display application version")
		version.Print()
		fmt.Println()
		applicationState.PrintConfig()
		lg.Debug("[MAIN] %s", logMsgCloseApp)
		os.Exit(0)
	}

	// If "-e" parameter provided, display enabled features and exit
	if applicationConfiguration.EnableFeature != "" {
		lg.Info("[MAIN] Enable feature: '%s'", applicationConfiguration.EnableFeature)
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
		lg.Info("[MAIN] Disable feature: '%s'", applicationConfiguration.DisableFeature)
		fmt.Printf("Disabled: '%s'\n", applicationConfiguration.DisableFeature)
		applicationState.SSHConfigEnabled = !(applicationConfiguration.DisableFeature == featureSSHConfig)
		err = applicationState.Persist()
		if err != nil {
			lg.Debug("[MAIN] Cannot save application configuration: %v", err)
		}

		lg.Debug("[MAIN] %s", logMsgCloseApp)
		os.Exit(0)
	}

	// Check errors in the very end. That allows to check application version and enable/disable
	// features, even if something is not right with the app.
	if !success {
		lg.Warn("--------= Close application with non-zero code =--------")
		os.Exit(1)
	}

	// Log application version
	lg.Info("[MAIN] Start application")
	lg.Info("[MAIN] Version:    %s", version.Number())
	lg.Info("[MAIN] Commit:     %s", version.CommitHash())
	lg.Info("[MAIN] Branch:     %s", version.BuildBranch())
	lg.Info("[MAIN] Build date: %s", version.BuildDate())

	return *applicationState
}

func createConfigurationOrExit() (application.Configuration, bool) {
	var fatalErr error
	success := true

	envConfig := application.Configuration{}
	// Parse environment parameters. These parameters have lower precedence than command line flags
	if err := env.Parse(&envConfig); err != nil {
		fmt.Printf("%+v\n", err)
	}

	// Check if "ssh" utility is in application path
	if err := utils.CheckAppInstalled("ssh"); err != nil {
		log.Fatalf("[MAIN] ssh utility is not installed or cannot be found in the executable path: %v", err)
	}

	cmdConfig := application.Configuration{}
	// Command line parameters have the highest precedence
	flag.BoolVar(&cmdConfig.DisplayVersionAndExit, "v", false, "Display application details")
	flag.StringVar(&cmdConfig.AppHome, "f", envConfig.AppHome, "Application home folder")
	flag.StringVar(&cmdConfig.LogLevel, "l", envConfig.LogLevel, "Log verbosity level: debug, info")
	flag.StringVar(&cmdConfig.SSHConfigFilePath, "s", envConfig.SSHConfigFilePath, "Specifies an alternative per-user SSH configuration file path")
	flag.Var(&cmdConfig.EnableFeature, "e", fmt.Sprintf("Enable feature. Supported values: %s", strings.Join(application.SupportedFeatures, "|")))
	flag.Var(&cmdConfig.DisableFeature, "d", fmt.Sprintf("Disable feature. Supported values: %s", strings.Join(application.SupportedFeatures, "|")))
	flag.Parse()

	// Set application home folder path
	cmdConfig.AppHome, fatalErr = utils.AppDir(appName, cmdConfig.AppHome)
	if fatalErr != nil {
		log.Printf("[MAIN] Can't set application home folder: %v\n", fatalErr)
		success = false
	}

	// Set ssh config file path
	cmdConfig.SSHConfigFilePath, fatalErr = utils.SSHConfigFilePath(cmdConfig.SSHConfigFilePath)
	if fatalErr != nil {
		log.Printf("[MAIN] Can't set SSH config path. Error: %v\n", fatalErr)
		success = false
	}

	// Create application folder
	if fatalErr = utils.CreateAppDirIfNotExists(cmdConfig.AppHome); fatalErr != nil {
		log.Printf("[MAIN] Can't create application home folder: %v\n", fatalErr)
		success = false
	}

	return cmdConfig, success
}
