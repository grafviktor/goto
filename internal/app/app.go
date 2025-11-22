package app

import (
	"os"

	"github.com/grafviktor/goto/internal/constant"
	"github.com/grafviktor/goto/internal/state"
	"github.com/grafviktor/goto/internal/storage"
	"github.com/grafviktor/goto/internal/ui"
	"github.com/grafviktor/goto/internal/ui/theme"
	"github.com/grafviktor/goto/internal/version"
)

func Start(st *state.State) error {
	st.Logger.Info("[APP] Start application")
	st.Logger.Debug("[APP] Parameters: %+v", os.Args[1:])

	version.LogDetails(st.Logger)
	st.LogDetails()

	var err error
	switch st.AppMode {
	case constant.AppModeType.StartUI:
		err = startUI(st)
	case constant.AppModeType.DisplayInfo:
		st.PrintConfig()
	case constant.AppModeType.HandleParam:
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

	err = theme.Load(st.AppHome, st.Theme, st.Logger)
	if err != nil {
		st.Logger.Error("[APP] Cannot load theme %q: %v. Fall back to default theme", st.Theme, err)
	}

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
