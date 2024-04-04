package ui

import (
	"context"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/grafviktor/goto/internal/state"
	"github.com/grafviktor/goto/internal/storage"
)

// Start - starts UI subsystem of the application. *state.ApplicationState should be substituted
// with interface type which would have getters and setters for appropriate fields, without doing it
// it's hard to use mock objects in unit tests of the child components. Search for 'MockAppState'.
func Start(ctx context.Context, storage storage.HostStorage, appState *state.ApplicationState, logger iLogger) {
	uiComponent := New(ctx, storage, appState, logger)
	p := tea.NewProgram(&uiComponent, tea.WithAltScreen())

	logger.Debug("[UI] Start user interface")
	if _, err := p.Run(); err != nil {
		logger.Error("[UI] Error starting user interface: %v", err)
		os.Exit(1)
	}
}
