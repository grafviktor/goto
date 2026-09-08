package sshsession

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/grafviktor/termview"
	"github.com/stretchr/testify/require"

	testutils "github.com/grafviktor/goto/internal/testutils"
)

func TestNew(t *testing.T) {
	_, err := New(80, 24, "echo", "test")
	require.NoError(t, err)
}

func TestNew_ExecutableNotFoundErr(t *testing.T) {
	_, err := New(80, 24, "no_such_binary", "test")
	require.Error(t, err)
}

func TestUpdate(t *testing.T) {
	t.Skip()

	tests := []struct {
		name            string
		sentMessage     tea.Msg
		expectedMessage tea.Msg
	}{{
		name:        "Handle termview.ClosedMsg",
		sentMessage: termview.ClosedMsg{},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, _ := New(80, 24, "echo", "test")
			_, cmd := m.Update(tt.sentMessage)
			var msgs []tea.Msg
			testutils.CmdToMessage(cmd, &msgs)

			require.IsType(t, tt.expectedMessage, msgs[0])
		})
	}
}
