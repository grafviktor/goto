// Package main contains application entry point
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/caarlos0/env/v10"

	"github.com/grafviktor/goto/internal/config"
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

const appName = "goto"

func main() {
	// Set application version and build details
	version.Set(buildVersion, buildCommit, buildBranch, buildDate)

	appState := createApplicationOrExit()
	appConfig := appState.ApplicationConfig

	// Logger created. Immediately print application version
	appConfig.Logger.Info("[MAIN] Start application")
	appConfig.Logger.Info("[MAIN] Version:    %s", version.Number())
	appConfig.Logger.Info("[MAIN] Commit:     %s", version.CommitHash())
	appConfig.Logger.Info("[MAIN] Branch:     %s", version.BuildBranch())
	appConfig.Logger.Info("[MAIN] Build date: %s", version.BuildDate())

	storage, fatalErr := storage.Get(appConfig.Context, appConfig, appConfig.Logger)
	if fatalErr != nil {
		appConfig.Logger.Error("[MAIN] Cannot access application storage: %v\n", fatalErr)
		os.Exit(1)
	}

	// Run user interface
	ui.Start(appConfig.Context, storage, &appState, appConfig.Logger)

	// Quit signal should be intercepted on the UI level, however it will require an
	// additional switch-case block with an appropriate checks. Leaving this message here.
	appConfig.Logger.Debug("[MAIN] Receive quit signal")
	appConfig.Logger.Debug("[MAIN] Save application state")
	fatalErr = appState.Persist()
	if fatalErr != nil {
		appConfig.Logger.Error("[MAIN] Can't save application state before closing %v", fatalErr)
	}

	appConfig.Logger.Info("[MAIN] Close application")
}

func createApplicationOrExit() state.Application {
	environmentParams := config.User{}
	// Parse environment parameters. These parameters have lower precedence than command line flags
	if err := env.Parse(&environmentParams); err != nil {
		fmt.Printf("%+v\n", err)
	}

	// Check if "ssh" utility is in application path
	if err := utils.CheckAppInstalled("ssh"); err != nil {
		log.Fatalf("[MAIN] ssh utility is not installed or cannot be found in the executable path: %v", err)
	}

	commandLineParams := config.User{}
	displayApplicationDetailsAndExit := false
	// Command line parameters have the highest precedence
	flag.BoolVar(&displayApplicationDetailsAndExit, "v", false, "Display application details")
	flag.StringVar(&commandLineParams.AppHome, "f", environmentParams.AppHome, "Application home folder")
	flag.StringVar(&commandLineParams.LogLevel, "l", environmentParams.LogLevel, "Log verbosity level: debug, info")
	flag.StringVar(&commandLineParams.SSHConfigFilePath, "s", environmentParams.SSHConfigFilePath, "Specifies an alternative per-user SSH configuration file path")
	flag.Var(&commandLineParams.EnableFeature, "e", fmt.Sprintf("Enable feature. Supported values: %s", strings.Join(config.SupportedFeatures, "|")))
	flag.Var(&commandLineParams.DisableFeature, "d", fmt.Sprintf("Disable feature. Supported values: %s", strings.Join(config.SupportedFeatures, "|")))
	flag.Parse()

	var fatalErr error
	// Set application home folder path
	commandLineParams.AppHome, fatalErr = utils.AppDir(appName, commandLineParams.AppHome)
	if fatalErr != nil {
		log.Printf("[MAIN] Can't set application home folder: %v\n", fatalErr)
	}

	// Set ssh config file path
	commandLineParams.SSHConfigFilePath, fatalErr = utils.SSHConfigFilePath(commandLineParams.SSHConfigFilePath)
	if fatalErr != nil {
		log.Printf("[MAIN] Can't set SSH config path. Error: %v\n", fatalErr)
	}

	// Create application folder
	if fatalErr = utils.CreateAppDirIfNotExists(commandLineParams.AppHome); fatalErr != nil {
		log.Printf("[MAIN] Can't create application home folder: %v\n", fatalErr)
	}

	// Create application logger
	lg, fatalErr := logger.New(commandLineParams.AppHome, commandLineParams.LogLevel)
	if fatalErr != nil {
		log.Printf("[MAIN] Can't create log file: %v\n", fatalErr)
	}

	userDefinedConfiguration := config.Merge(environmentParams, commandLineParams, &lg)

	// Create applicationConfiguration state
	applicationConfiguration := config.New(context.Background(), userDefinedConfiguration, &lg)
	applicationState := state.Create(applicationConfiguration, &lg)

	// If "-v" parameter provided, display application version configuration and exit
	if displayApplicationDetailsAndExit {
		lg.Debug("[MAIN] Display application version")
		version.Print()
		fmt.Println()
		applicationState.PrintConfig()

		lg.Debug("[MAIN] Exit application")
		os.Exit(0)
	}

	if fatalErr != nil {
		lg.Error("[MAIN] Fatal error:", fatalErr)
		os.Exit(1)
	}

	// If "-e" parameter provided, display enabled features and exit
	if userDefinedConfiguration.EnableFeature != "" {
		lg.Debug("[MAIN] Enable feature: '%s'", userDefinedConfiguration.EnableFeature)
		fmt.Printf("Enabled: '%s'\n", userDefinedConfiguration.EnableFeature)
		applicationState.SSHConfigEnabled = userDefinedConfiguration.EnableFeature == "ssh_config"
		applicationState.Persist()

		lg.Debug("[MAIN] Exit application")
		os.Exit(0)
	}

	// If "-d" parameter provided, display disabled features and exit
	if userDefinedConfiguration.DisableFeature != "" {
		lg.Debug("[MAIN] Disable feature: '%s'", userDefinedConfiguration.DisableFeature)
		fmt.Printf("Disabled: '%s'\n", userDefinedConfiguration.DisableFeature)
		applicationState.SSHConfigEnabled = !(userDefinedConfiguration.DisableFeature == "ssh_config")
		applicationState.Persist()

		lg.Debug("[MAIN] Exit application")
		os.Exit(0)
	}

	// config.application, state.application
	// return application, *appState
	return *applicationState
}
