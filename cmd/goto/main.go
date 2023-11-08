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

func main() {
	// Set application version and build details
	version.Set(buildVersion, buildDate, buildCommit)

	var environmentSettings config.EnvironmentSettings

	if err := env.Parse(&environmentSettings); err != nil {
		fmt.Printf("%+v\n", err)
	}

	appConfig := config.New()

	flag.StringVar(&appConfig.ConfigFile, "c", environmentSettings.ConfigFile, "Application configuration file")
	flag.StringVar(&appConfig.HostsFilePath, "f", environmentSettings.HostsFilePath, "Path to yaml file with hosts")
	flag.StringVar(&appConfig.LogLevel, "l", environmentSettings.LogLevel, "Log level")
	flag.Parse()

	// fileConfig, ok := config.ReadFromFile(appConfig.ConfigFilePath)

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
