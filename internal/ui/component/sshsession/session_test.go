package sshsession

import (
	"errors"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/grafviktor/termview"
	"github.com/stretchr/testify/require"

	testutils "github.com/grafviktor/goto/internal/testutils"
	"github.com/grafviktor/goto/internal/ui/message"
	"github.com/grafviktor/goto/internal/utils"
)

func TestNew(t *testing.T) {
	newTermView = func(opts ...termview.Option) (termview.Model, error) {
		return termview.Model{}, nil
	}
	t.Cleanup(func() { newTermView = termview.New })

	m, err := New(80, 24, "echo", "test")
	require.NoError(t, err)
	require.Equal(t, "echo", m.command)
}

func TestNew_ExecutableNotFoundErr(t *testing.T) {
	newTermView = func(opts ...termview.Option) (termview.Model, error) {
		return termview.Model{}, errors.New("executable file not found")
	}
	t.Cleanup(func() { newTermView = termview.New })

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
			m := Model{stdErr: &utils.ProcessBufferWriter{}}
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
