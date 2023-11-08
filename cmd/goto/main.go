package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/caarlos0/env/v10"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/grafviktor/goto/internal/config"
	"github.com/grafviktor/goto/internal/logger"
	"github.com/grafviktor/goto/internal/state"
	"github.com/grafviktor/goto/internal/storage"
	"github.com/grafviktor/goto/internal/ui"
	"github.com/grafviktor/goto/internal/version"
)

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func readSettings() config.UserSettings {

	var userSettings config.UserSettings

	if err := env.Parse(&userSettings); err != nil {
		fmt.Printf("%+v\n", err)
	}

	// Read environment variables, fallback to file settings
	flag.StringVar(&userSettings.HostsFilePath, "f", userSettings.HostsFilePath, "Path to yaml file with the list of hosts")
	flag.StringVar(&userSettings.LogFilePath, "p", userSettings.LogFilePath, "Log file path")
	flag.StringVar(&userSettings.LogLevel, "l", userSettings.LogLevel, "Log level: debug, info, none")
	flag.Parse()
}

func main() {
	// Set application version and build details
	version.Set(buildVersion, buildDate, buildCommit)

	lg, err := logger.New("goto", logger.LevelDebug)
	if err != nil {
		log.Fatalf("Can't create log file %v", err)
	}

	lg.Info("Starting application")
	lg.Debug("Version %s", version.BuildVersion())
	lg.Debug("Build date %s", version.BuildDate())
	lg.Debug("Commit %s", version.BuildCommit())

	ctx := context.Background()
	appConfig := config.New(ctx, &lg)

	hostStorage, err := storage.Get(ctx, appConfig)
	if err != nil {
		lg.Debug("Error running program: %v", err)
		os.Exit(1)
	}

	appState := state.Get(appConfig.HomeFolder, &lg)
	uiComponent := ui.NewMainModel(ctx, hostStorage, appState, &lg)
	p := tea.NewProgram(&uiComponent, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		lg.Debug("Error running program: %v", err)
		os.Exit(1)
	}

	lg.Debug("Save application state")
	err = appState.Persist()
	if err != nil {
		lg.Debug("Can't save application state before closing %v", err)
	}

	lg.Debug("Close the application")
}
