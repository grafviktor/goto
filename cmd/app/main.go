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
)

func WithLogger(ctx context.Context, logger *logger.Logger) context.Context {
	return context.WithValue(ctx, "log", logger)
}

func main() {
	ctx := context.Background()
	ctxWithLogger := logger.ToContext(ctx, &logger.Logger{})
	conf := config.Application{
		AppName: "goto",
		Context: ctxWithLogger,
	}

	st, err := storage.GetStorage(ctx, conf)
	if err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}

	logger := logger.Logger{}

	// var logFile *os.File
	// if len(os.Getenv("DEBUG")) > 0 || true { // TODO: remove force debug flag
	// 	logFile, err = tea.LogToFile("debug.log", "debug")
	// 	if err != nil {
	// 		fmt.Println("fatal:", err)
	// 		os.Exit(1)
	// 	}
	// }

	// log.Println("Starting application")
	logger.Log("Starting application")

	uiComponent := ui.NewMainModel(ctx, conf, st)
	p := tea.NewProgram(uiComponent, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		log.Println("Error running program:", err)

		os.Exit(1)
	}
}
