package ui

import (
	"context"
	"errors"
	"syscall"

	tea "charm.land/bubbletea/v2"

	"github.com/grafviktor/goto/internal/state"
	"github.com/grafviktor/goto/internal/storage"
)

// Start - starts UI subsystem of the application. *state.ApplicationState should be substituted
// with interface type which would have getters and setters for appropriate fields, without doing it
// it's hard to use mock objects in unit tests of the child components. Search for 'MockAppState'.
func Start(ctx context.Context, storage storage.HostStorage, appState *state.State) error {
	if ctx.Err() != nil {
		// I use it in tests to prevent UI start
		return ctx.Err()
	}

	uiComponent := New(ctx, storage, appState, appState.Logger)
	p := tea.NewProgram(&uiComponent)

	appState.Logger.Debug("[UI] Start user interface")
	if _, err := p.Run(); err != nil {
		var errno syscall.Errno
		if errors.As(err, &errno) {
			appState.Logger.Error("[UI] Syscall error code: %v %T %#v", err, err, err)
			return errors.New("unsupported terminal type or terminal is running in legacy mode")
		}

		appState.Logger.Error("[UI] Error starting user interface: %v", err)
		return err
	}

	// There is no way to return error from Bubble Tea application,
	// so we need to read error right from the model object to check
	// if application closed with error or not.
	if uiComponent.exitError != nil {
		appState.Logger.Error("[UI] Application closed with error: %v", uiComponent.exitError)
		return uiComponent.exitError
	}

	return nil
}
