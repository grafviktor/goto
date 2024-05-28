// Package main contains application entry point
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

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
	flag.Parse()

	var err error
	// Get application home folder path
	commandLineParams.AppHome, err = utils.AppDir(appName, commandLineParams.AppHome)
	if err != nil {
		log.Fatalf("[MAIN] Can't get application home folder: %v", err)
	}

	// Create application folder
	if err = utils.CreateAppDirIfNotExists(commandLineParams.AppHome); err != nil {
		log.Fatalf("[MAIN] Can't create application home folder: %v", err)
	}

	// Create application logger
	lg, err := logger.New(commandLineParams.AppHome, commandLineParams.LogLevel)
	if err != nil {
		log.Fatalf("[MAIN] Can't create log file: %v", err)
	}

	// Create application configuration and set application home folder
	appConfig := config.Merge(environmentParams, commandLineParams, &lg)

	// If "-v" parameter provided, display application version configuration and exit
	if displayApplicationDetailsAndExit {
		lg.Debug("[MAIN] Display application version")
		version.Print()
		fmt.Println()
		appConfig.Print()

		lg.Debug("[MAIN] Exit application")
		os.Exit(0)
	}

	// Logger created. Immediately print application version
	lg.Info("[MAIN] Start application")
	lg.Info("[MAIN] Version:    %s", version.Number())
	lg.Info("[MAIN] Commit:     %s", version.CommitHash())
	lg.Info("[MAIN] Branch:     %s", version.BuildBranch())
	lg.Info("[MAIN] Build date: %s", version.BuildDate())

	// Create application state
	ctx := context.Background()
	application := config.NewApplication(ctx, appConfig, &lg)
	appState := state.Get(application.Config.AppHome, &lg)
	storage, err := storage.Get(ctx, application)
	if err != nil {
		lg.Error("[MAIN] Error running application: %v", err)
		os.Exit(1)
	}

	// Run user interface
	ui.Start(ctx, storage, appState, &lg)

	// Quit signal should be intercepted on the UI level, however it will require an
	// additional switch-case block with an appropriate checks. Leaving this message here.
	lg.Debug("[MAIN] Receive quit signal")
	lg.Debug("[MAIN] Save application state")
	err = appState.Persist()
	if err != nil {
		lg.Error("[MAIN] Can't save application state before closing %v", err)
	}

	lg.Info("[MAIN] Close application")
}
