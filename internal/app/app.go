package app

import (
	"os"

	"github.com/grafviktor/goto/internal/config"
	"github.com/grafviktor/goto/internal/state"
	"github.com/grafviktor/goto/internal/storage"
	"github.com/grafviktor/goto/internal/ui"
	"github.com/grafviktor/goto/internal/version"
)

func Start(st *state.State) error {
	st.Logger.Info("[APP] Start application")
	st.Logger.Debug("[APP] Parameters: %+v", os.Args[1:])

	version.LogDetails(st.Logger)
	st.LogDetails(st.Logger)

	var err error
	switch st.AppMode {
	case config.AppModeType.StartUI:
		err = startUI(st)
	case config.AppModeType.DisplayInfo:
		st.PrintConfig()
	default:
		// nop - proceed to exit
	}

	return err
}

func startUI(st *state.State) error {
	// Init storage
	str, err := storage.Initialize(st.Context, st, st.Logger)
	if err != nil {
		return err
	}

	defer func() {
		st.Logger.Debug("[APP] Close storage")
		str.Close()
	}()

	// Initialize themes
	// theme.Initialize(st.Theme, st.AppHome, st.Logger)

	// Run user interface and block
	err = ui.Start(st.Context, str, st)
	if err != nil {
		return err
	}

	// Quit signal should be intercepted on the UI level, however it will require
	// additional switch-case block with appropriate checks. Leaving this message here.
	st.Logger.Debug("[APP] Receive quit signal")

	return nil
}
