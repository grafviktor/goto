package app

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/grafviktor/goto/internal/constant"
	"github.com/grafviktor/goto/internal/state"
	"github.com/grafviktor/goto/internal/testutils/mocklogger"
)

func Test_Start(t *testing.T) {
	tmpDir := t.TempDir()
	logger := mocklogger.Logger{}
	// To prevent UI start, we use already cancelled context
	// otherwise the test would block waiting for user input
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	st := state.State{
		Context: ctx,
		Logger:  &logger,
		AppHome: tmpDir,
	}

	tests := []struct {
		name    string
		AppMode constant.AppMode
		wantErr bool
	}{
		{
			name:    "Start UI mode",
			AppMode: constant.AppModeType.StartUI,
			wantErr: true, // it'll fail because /dev/tty isn't available
		},
		{
			name:    "Display Info mode",
			AppMode: constant.AppModeType.DisplayInfo,
			wantErr: false,
		},
		{
			name:    "Handle Param mode",
			AppMode: constant.AppModeType.HandleParam,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st.AppMode = tt.AppMode
			err := Start(&st)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func Test_startUI(t *testing.T) {
	// To prevent UI start, we use already cancelled context
	// otherwise the test would block waiting for user input
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	logger := mocklogger.Logger{}
	st := state.State{
		Context: ctx,
		Logger:  &logger,
		AppHome: t.TempDir(),
		Theme:   "nord",
	}

	err := startUI(&st)
	require.ErrorContains(t, err, "context canceled")
}
