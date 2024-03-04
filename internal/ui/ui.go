package ui

import (
	"context"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/grafviktor/goto/internal/state"
	"github.com/grafviktor/goto/internal/storage"
)

type interfaceLogger interface {
	Debug(format string, args ...any)
	Info(format string, args ...any)
	Error(format string, args ...any)
}

// Start - starts UI subsystem of the application.
func Start(ctx context.Context, storage storage.HostStorage, appState *state.ApplicationState, logger interfaceLogger) {
	uiComponent := New(ctx, storage, appState, logger)
	p := tea.NewProgram(&uiComponent, tea.WithAltScreen())

	logger.Debug("[UI] Start user interface")
	if _, err := p.Run(); err != nil {
		logger.Error("[UI] Error starting user interface: %v", err)
		os.Exit(1)
	}
}
