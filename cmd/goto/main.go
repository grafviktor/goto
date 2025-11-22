// Package main contains the application entry point for the GOTO SSH Manager.
//
//nolint:lll,gochecknoglobals // Disable line length check, Ignore burn in these variables.
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/grafviktor/goto/internal/app"
	"github.com/grafviktor/goto/internal/config"
	"github.com/grafviktor/goto/internal/constant"
	"github.com/grafviktor/goto/internal/logger"
	"github.com/grafviktor/goto/internal/state"
	"github.com/grafviktor/goto/internal/utils"
	"github.com/grafviktor/goto/internal/version"
)

var (
	buildVersion string
	buildDate    string
	buildCommit  string
	buildBranch  string
)

func main() {
	// Set application version and build details
	version.Set(buildVersion, buildCommit, buildBranch, buildDate)

	// Create application configuration
	cfg, err := config.Initialize()
	if err != nil {
		fmt.Printf("[MAIN] Error: %v\n", err)
		os.Exit(1)
	}

	// Check prerequisites
	err = utils.CheckAppRequirements(cfg.AppHome)
	if err != nil {
		fmt.Printf("[MAIN] Error: %v\n", err)
		os.Exit(1)
	}

	// Create application logger
	lgr, err := logger.Initialize(cfg.AppHome, cfg.LogLevel)
	if err != nil {
		fmt.Printf("[MAIN] Error: %v\n", err)
		os.Exit(1)
	}

	// Create state
	st, err := state.Initialize(context.Background(), cfg, lgr)
	if err != nil {
		logMessage := fmt.Sprintf("[CONFIG] Error: %v", err)
		utils.LogAndCloseApp(lgr, constant.AppExitCodeError, logMessage)
	}

	// Start application
	err = app.Start(st)
	if err != nil {
		logMessage := fmt.Sprintf("[MAIN] Error: %v", err)
		utils.LogAndCloseApp(lgr, constant.AppExitCodeError, logMessage)
	}

	lgr.Debug("[MAIN] Save application state")
	if err = st.Persist(); err != nil {
		logMessage := fmt.Sprintf("[MAIN] Can't save application state before closing: %v", err)
		utils.LogAndCloseApp(lgr, constant.AppExitCodeError, logMessage)
	}

	utils.LogAndCloseApp(lgr, constant.AppExitCodeSuccess, "")
}
