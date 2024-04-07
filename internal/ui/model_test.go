package ui

import (
	"context"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/grafviktor/goto/internal/mock"
	"github.com/grafviktor/goto/internal/state"
	"github.com/grafviktor/goto/internal/test"
	"github.com/grafviktor/goto/internal/ui/message"
	"github.com/grafviktor/goto/internal/utils/ssh"
)

func TestNew(t *testing.T) {
	model := New(context.TODO(), mock.NewMockStorage(true), MockAppState(), &mock.MockLogger{})
	require.NotNil(t, model)
}

func TestUpdate_KeyMsg(t *testing.T) {
	// Random key test - make sure that the app reacts on Ctrl+C
	model := New(context.TODO(), mock.NewMockStorage(true), MockAppState(), &mock.MockLogger{})
	_, cmd := model.Update(tea.KeyMsg{
		Type: tea.KeyCtrlC,
	})

	assert.NotNil(t, model)
	require.IsType(t, tea.QuitMsg{}, cmd(), "Wrong message type")
}

func TestUpdate_TerminalSizePolling(t *testing.T) {
	// Ensure that when the model receives TerminalSizePolling it autogenerates 'WindowSizeMsg'
	model := New(context.TODO(), mock.NewMockStorage(true), MockAppState(), &mock.MockLogger{})
	assert.Equal(t, 0, model.appState.Width)
	assert.Equal(t, 0, model.appState.Height)

	_, cmds := model.Update(message.TerminalSizePolling{
		Width:  10,
		Height: 10,
	})

	var dst []tea.Msg
	test.CmdToMessage(cmds, &dst)

	require.Contains(t, dst, tea.WindowSizeMsg{
		Width:  10,
		Height: 10,
	})
}

func MockAppState() *state.ApplicationState {
	return &state.ApplicationState{
		HostSSHConfig: &ssh.Config{},
	}
}
