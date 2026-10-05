package sshsession

import (
	"errors"
	"reflect"
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

func TestUpdate_MouseMsg(t *testing.T) {
	// Test that if we click on a position outside the terminal area, the mouse event is ignored.
	m, _ := New(80, 24, "hostname")
	// Remember, the terminal height is lower, than the whole component height, because of the
	// status line.
	require.Equal(t, 22, m.term.Height())
	// Now click one pixel below the terminal and start dragging the one pixel right.
	// Mouse release event will be ignored, because it is outside the terminal area.
	m.Update(tea.MouseClickMsg{Button: tea.MouseLeft, X: 10, Y: 23})
	m.Update(tea.MouseMotionMsg{Button: tea.MouseLeft, X: 11, Y: 23})
	_, cmd := m.Update(tea.MouseReleaseMsg{Button: tea.MouseLeft, X: 11, Y: 23})
	// It'll return nil
	require.Nil(t, cmd)
	// Now click on the last row of the terminal and start dragging the one pixel right.
	m.Update(tea.MouseClickMsg{Button: tea.MouseLeft, X: 10, Y: 22})
	m.Update(tea.MouseMotionMsg{Button: tea.MouseLeft, X: 11, Y: 22})
	_, cmd = m.Update(tea.MouseReleaseMsg{Button: tea.MouseLeft, X: 11, Y: 22})
	// This time, the message will be handled by the termview and we should get non-nil cmd
	require.NotNil(t, cmd)
}

func TestUpdate_HandleTextSelectedMsg(t *testing.T) {
	// Test that if message has a different ID it is ignored and the notification area text is not changed.
	text := `
A long time ago, in a galaxy far, far, away...

A vast sea of stars serves as the backdrop for the main title.
War drums echo through the heavens as a rollup slowly crawls into infinity.
It is a period of civil war. Rebel spaceships, striking from a
hidden base, have won their first victory against the evil Galactic Empire.
During the battle, Rebel spies managed to steal secret plans to the Empire's
ultimate weapon, the DEATH STAR...
`

	var cmd tea.Cmd
	m, _ := New(80, 24, "hostname")
	_, cmd = m.Update(termview.TextSelectedMsg{
		ID:   m.term.ID() + 1, // Not the same ID, so it should be ignored.
		Text: text,
	})
	require.Equal(t, m.defaultNotificationAreaText, m.currentNotificationAreaText)
	require.Nil(t, cmd)

	// Now use the same terminal ID, this time the text should be handled.
	_, cmd = m.Update(termview.TextSelectedMsg{
		ID:   m.term.ID(),
		Text: text,
	})
	require.Equal(t, "10 line(s) copied to clipboard", m.currentNotificationAreaText)
	var msgs []tea.Msg
	testutils.CmdToMessage(cmd, &msgs)
	require.Len(t, msgs, 2)
	require.Equal(t, "setClipboardMsg", reflect.TypeOf(msgs[0]).Name())
	require.Equal(t, "clearStatusMsg", reflect.TypeOf(msgs[1]).Name())
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
