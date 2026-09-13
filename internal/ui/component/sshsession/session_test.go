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

func TestHandleSessionClose(t *testing.T) {
	m, _ := New(80, 24, "hostname")
	cmd := m.handleSessionClose(termview.ClosedMsg{})
	require.IsType(t, message.RunProcessSuccess{}, cmd())

	cmd = m.handleSessionClose(termview.ClosedMsg{ProcessError: errors.New("mock error"), ProcessExitCode: 1})
	errorDetails := cmd()
	require.IsType(t, message.RunProcessErrorOccurred{}, errorDetails)
	require.Equal(t, "mock error", errorDetails.(message.RunProcessErrorOccurred).StdErr)
}
