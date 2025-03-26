package ui

import (
	"context"
	"os"
	"reflect"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/grafviktor/goto/internal/constant"
	"github.com/grafviktor/goto/internal/model/ssh"
	"github.com/grafviktor/goto/internal/state"
	testutils "github.com/grafviktor/goto/internal/testutils"
	"github.com/grafviktor/goto/internal/ui/message"
	"github.com/grafviktor/goto/internal/utils"
)

func TestNew(t *testing.T) {
	model := New(context.TODO(), testutils.NewMockStorage(false), MockAppState(), &testutils.MockLogger{})
	require.NotNil(t, model)
	cmd := model.Init()
	var msgs []tea.Msg
	testutils.CmdToMessage(cmd, &msgs)

	require.IsType(t, message.HostSelected{}, msgs[0])
	require.IsType(t, message.RunProcessSSHLoadConfig{}, msgs[1])
}

func TestUpdate_KeyMsg(t *testing.T) {
	// Random key test - make sure that the app reacts on Ctrl+C
	model := New(context.TODO(), testutils.NewMockStorage(true), MockAppState(), &testutils.MockLogger{})
	_, cmd := model.Update(tea.KeyMsg{
		Type: tea.KeyCtrlC,
	})

	assert.NotNil(t, model)
	require.IsType(t, tea.QuitMsg{}, cmd(), "Wrong message type")
}

func TestDispatchProcess_Foreground(t *testing.T) {
	// Create a model
	model := New(context.TODO(), testutils.NewMockStorage(true), MockAppState(), &testutils.MockLogger{})

	validProcess := utils.BuildProcess("echo test") // "echo test" is a cross-platform command
	validProcess.Stdout = os.Stdout
	validProcess.Stderr = &utils.ProcessBufferWriter{}
	teaCmd := model.dispatchProcess("test", validProcess, false, false)

	// Perform assertions
	assert.Equal(t, reflect.Func, reflect.ValueOf(teaCmd).Kind())

	// Unwrap tea.execMsg
	execMsg := teaCmd()
	assert.Equal(t, "tea.execMsg", reflect.TypeOf(execMsg).String())

	// Get access to callback function from tea.execMsg
	callbackFn := reflect.ValueOf(execMsg).FieldByName("fn")
	assert.Equal(t, reflect.Func, callbackFn.Kind())

	// This is the dead end because reflection does not allow us to call private methods
	// ```
	// type execMsg struct {
	// 	cmd ExecCommand  // private
	// 	fn  ExecCallback // private
	// }
	// ```
	// argVals := make([]reflect.Value, 1)
	// argVals[0] = reflect.ValueOf(errors.New("Mock Error"))
	// callbackFn.Call(argVals)
}

// This test is failing in a real Windows environment with error 'exec: "echo": executable file not found in %PATH%'.
// Low priority though as it works in gitlab tests for Windows platform. Requires investigation.
func TestDispatchProcess_Background_OK(t *testing.T) {
	// Create a model
	model := New(context.TODO(), testutils.NewMockStorage(true), MockAppState(), &testutils.MockLogger{})

	// Test case: Successful process execution
	validProcess := utils.BuildProcess("echo test")
	validProcess.Stdout = &utils.ProcessBufferWriter{}
	validProcess.Stderr = &utils.ProcessBufferWriter{}

	callbackFnResult := model.dispatchProcess("test", validProcess, true, false)

	// Perform assertions
	assert.Equal(t, reflect.Func, reflect.ValueOf(callbackFnResult).Kind())

	result := callbackFnResult()
	require.IsType(t, message.RunProcessSuccess{}, result)
}

func TestDispatchProcess_Background_Fail(t *testing.T) {
	// Create a model
	model := New(context.TODO(), testutils.NewMockStorage(true), MockAppState(), &testutils.MockLogger{})

	// Test case: Unsuccessful process execution
	validProcess := utils.BuildProcess("nonexistent command")
	validProcess.Stdout = &utils.ProcessBufferWriter{}
	validProcess.Stderr = &utils.ProcessBufferWriter{}

	callbackFnResult := model.dispatchProcess("test", validProcess, true, false)

	// Perform assertions
	assert.Equal(t, reflect.Func, reflect.ValueOf(callbackFnResult).Kind())

	result := callbackFnResult()
	require.IsType(t, message.RunProcessErrorOccurred{}, result)
}

func TestHandleProcessSuccess_SSH_load_config(t *testing.T) {
	model := New(context.TODO(), testutils.NewMockStorage(true), MockAppState(), &testutils.MockLogger{})
	given := message.RunProcessSuccess{
		ProcessType: constant.ProcessTypeSSHLoadConfig,
		StdOut:      "hostname localhost\r\nport 2222\r\nidentityfile /tmp\r\nuser root",
	}

	expected := message.HostSSHConfigLoaded{
		HostID: 0,
		Config: ssh.Config{
			Hostname:     "localhost",
			IdentityFile: "/tmp",
			Port:         "2222",
			User:         "root",
		},
	}

	actual := model.handleProcessSuccess(given)()
	require.Equal(t, expected, actual)
}

func TestHandleProcessSuccess_SSH_copy_ID(t *testing.T) {
	type expected struct {
		modelMessage string
		viewState    int
	}

	tests := []struct {
		name     string
		msg      message.RunProcessSuccess
		expected expected
	}{
		{
			name: "Handle SSH copy ID when process output contains an error",
			msg: message.RunProcessSuccess{
				ProcessType: constant.ProcessTypeSSHCopyID,
				StdOut:      "normal output",
				StdErr:      "foo ERROR bar",
			},
			expected: expected{
				modelMessage: "foo ERROR bar",
				viewState:    (int)(state.ViewMessage),
			},
		},
		{
			name: "Handle SSH copy ID when process output contains a warning message",
			msg: message.RunProcessSuccess{
				ProcessType: constant.ProcessTypeSSHCopyID,
				StdOut:      "normal output",
				StdErr:      "foo WARNING bar",
			},
			expected: expected{
				modelMessage: "foo WARNING bar",
				viewState:    (int)(state.ViewMessage),
			},
		},
		{
			name: "Handle SSH copy ID when it ended successfully",
			msg: message.RunProcessSuccess{
				ProcessType: constant.ProcessTypeSSHCopyID,
				StdOut:      "normal output",
				StdErr:      "foo SOMETHING bar",
			},
			expected: expected{
				modelMessage: "normal output",
				viewState:    (int)(state.ViewMessage),
			},
		},
		{
			name: "Unsupported process",
			msg: message.RunProcessSuccess{
				ProcessType: constant.ProcessTypeSSHConnect,
				StdOut:      "normal output",
				StdErr:      "foo SOMETHING bar",
			},
			expected: expected{
				modelMessage: "",
				viewState:    (int)(state.ViewHostList),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := New(context.TODO(), testutils.NewMockStorage(true), MockAppState(), &testutils.MockLogger{})
			model.handleProcessSuccess(tt.msg)
			require.Equal(t, tt.expected.modelMessage, model.viewMessageContent)
			require.Equal(t, tt.expected.viewState, (int)(model.appState.CurrentView))
		})
	}
}

func TestHandleProcessError(t *testing.T) {
	type expected struct {
		modelMessage string
		viewState    int
	}

	tests := []struct {
		name     string
		msg      message.RunProcessErrorOccurred
		expected expected
	}{
		{
			name: "Handle process error stderr and stdout are populated",
			msg: message.RunProcessErrorOccurred{
				StdOut: "normal output",
				StdErr: "error output",
			},
			expected: expected{
				modelMessage: "error output\nDetails: normal output",
				viewState:    (int)(state.ViewMessage),
			},
		},
		{
			name: "Handle process error only stderr is populated",
			msg: message.RunProcessErrorOccurred{
				StdErr: "error output",
			},
			expected: expected{
				modelMessage: "error output",
				viewState:    (int)(state.ViewMessage),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := New(context.TODO(), testutils.NewMockStorage(true), MockAppState(), &testutils.MockLogger{})
			model.handleProcessError(tt.msg)
			require.Equal(t, tt.expected.modelMessage, model.viewMessageContent)
			require.Equal(t, tt.expected.viewState, (int)(model.appState.CurrentView))
		})
	}
}

// ---------------------------------

func MockAppState() *state.ApplicationState {
	return &state.ApplicationState{}
}
