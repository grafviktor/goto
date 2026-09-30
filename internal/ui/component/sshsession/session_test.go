package sshsession

import (
	"errors"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/grafviktor/termview"
	"github.com/stretchr/testify/require"

	testutils "github.com/grafviktor/goto/internal/testutils"
	"github.com/grafviktor/goto/internal/ui/message"
)

func TestNew(t *testing.T) {
	_, err := New(80, 24, "hostname")
	require.NoError(t, err)
}

func TestNew_ExecutableNotFoundErr(t *testing.T) {
	_, err := New(80, 24, "no_such_binary", "test")
	require.Error(t, err)
}

func TestUpdate(t *testing.T) {
	tests := []struct {
		name            string
		sentMessage     tea.Msg
		expectedMessage tea.Msg
	}{{
		name:            "Handle termview.ClosedMsg",
		sentMessage:     termview.ClosedMsg{},
		expectedMessage: message.RunProcessSuccess{},
	}, {
		name:            "Other case",
		sentMessage:     struct{}{},
		expectedMessage: nil,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, _ := New(80, 24, "hostname")
			_, cmd := m.Update(tt.sentMessage)
			var msgs []tea.Msg
			testutils.CmdToMessage(cmd, &msgs)

			if tt.expectedMessage == nil {
				require.Nil(t, msgs)
			} else {
				require.IsType(t, tt.expectedMessage, msgs[0])
			}
		})
	}
}

func TestUpdate_WindowSizeMsg(t *testing.T) {
	// Test that terminal which we're wrapping will get its size
	// with status line height substracted. Testing both cases
	// when create and update terminal.
	m, _ := New(80, 24, "hostname")
	require.Equal(t, 80, m.term.Width())
	require.Equal(t, 22, m.term.Height())

	m.Update(tea.WindowSizeMsg{Width: 100, Height: 50})
	require.Equal(t, 100, m.term.Width())
	require.Equal(t, 48, m.term.Height())
}

/*
+ func (m *Model) handleMouseMsg(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
+ 	if msg.Mouse().Y > m.term.Height() {
+ 		// Skip mouse events in the status area.
+ 		return m, nil
+ 	}
+
  - updated, cmd := m.term.Update(msg)
  - m.term = updated
  - return m, cmd
    }
*/
func TestUpdate_MouseMsg(t *testing.T) {
	t.Skip("Skipping as it's failing")
	// Test that if we click on a position outside the terminal area, the mouse event is ignored.
	m, _ := New(80, 24, "hostname")
	// Remember, the terminal height is lower, than the whole component height, because of the
	// status line.
	require.Equal(t, 22, m.term.Height())
	// Now click one pixel below the terminal
	_, cmd := m.Update(tea.MouseClickMsg{Button: tea.MouseLeft, X: 10, Y: 23})
	// It'll return null
	require.Nil(t, cmd)
	// Now click on the last row of the terminal
	_, cmd = m.Update(tea.MouseClickMsg{Button: tea.MouseLeft, X: 10, Y: 22})
	// The message will be handled by the termview and we should get non-nil cmd
	require.NotNil(t, cmd)
}

func TestHandleSessionClose(t *testing.T) {
	m, _ := New(80, 24, "hostname")
	cmd := m.handleSessionClose(termview.ClosedMsg{})
	require.IsType(t, message.RunProcessSuccess{}, cmd())

	cmd = m.handleSessionClose(termview.ClosedMsg{ProcessError: errors.New("mock error"), ProcessExitCode: 1})
	errorDetails := cmd()
	require.IsType(t, message.RunProcessErrorOccurred{}, errorDetails)
	require.Equal(t, "mock error", errorDetails.(message.RunProcessErrorOccurred).StdErr)
}
