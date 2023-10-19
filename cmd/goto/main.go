package main

import (
	"context"
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/grafviktor/goto/internal/config"
	"github.com/grafviktor/goto/internal/logger"
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
	lg := logger.Logger{}
	// Set application version and build details
	version.Set(buildVersion, buildDate, buildCommit)
	lg.Log("Starting application")
	lg.Log("Version %s", version.BuildVersion())
	lg.Log("Build date %s", version.BuildDate())
	lg.Log("Commit %s", version.BuildCommit())

	ctx := context.Background()
	ctxWithLogger := logger.ToContext(ctx, &lg)
	conf := config.Application{
		AppName: "goto",
		Context: ctxWithLogger,
	}

	st, err := storage.GetStorage(ctx, conf)
	if err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}

	uiComponent := ui.NewMainModel(ctx, conf, st)
	p := tea.NewProgram(uiComponent, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		log.Println("Error running program:", err)

		os.Exit(1)
	}
}
