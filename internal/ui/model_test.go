package ui

import (
	"context"
	"errors"
	"os"
	"reflect"
	"runtime"
	"testing"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	uv "github.com/charmbracelet/ultraviolet"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/grafviktor/goto/internal/constant"
	hostModel "github.com/grafviktor/goto/internal/model/host"
	"github.com/grafviktor/goto/internal/model/sshconfig"
	"github.com/grafviktor/goto/internal/state"
	testutils "github.com/grafviktor/goto/internal/testutils"
	"github.com/grafviktor/goto/internal/testutils/mocklogger"
	"github.com/grafviktor/goto/internal/ui/message"
	"github.com/grafviktor/goto/internal/utils"
)

func TestNew(t *testing.T) {
	model := New(context.TODO(), testutils.NewMockStorage(false), MockAppState(), &mocklogger.Logger{})
	require.NotNil(t, model)
	cmd := model.Init()
	var msgs []tea.Msg
	testutils.CmdToMessage(cmd, &msgs)

	require.IsType(t, message.HostSelect{}, msgs[0])
	require.IsType(t, message.RunProcessSSHLoadConfig{}, msgs[1])
}

func TestUpdate_KeyMsg(t *testing.T) {
	// Random key test - make sure that the app reacts on Ctrl+C
	model := New(context.TODO(), testutils.NewMockStorage(true), MockAppState(), &mocklogger.Logger{})
	_, cmd := model.Update(tea.KeyPressMsg{
		Code: 'c',
		Mod:  tea.ModCtrl,
	})

	assert.NotNil(t, model)
	require.IsType(t, tea.QuitMsg{}, cmd(), "Wrong message type")
}

func TestUpdate_WindowSizeMsg(t *testing.T) {
	model := New(context.TODO(), testutils.NewMockStorage(false), MockAppState(), &mocklogger.Logger{})
	require.Zero(t, model.appState.Width)
	require.Zero(t, model.appState.Height)
	require.False(t, model.ready)

	model.Update(tea.WindowSizeMsg{
		Width:  100,
		Height: 50,
	})

	require.Equal(t, 100, model.appState.Width)
	require.Equal(t, 50, model.appState.Height)
	require.Equal(t, 100, model.viewport.Width())
	require.Equal(t, 50, model.viewport.Height())
	require.True(t, model.ready)
}

func TestDispatchProcess_SSH_connect(t *testing.T) {
	tests := []struct {
		name             string
		embeddedTerminal bool
		expectedMsgName  string
	}{
		{
			name:             "Process SSH connect message with embedded terminal",
			embeddedTerminal: true,
			expectedMsgName:  "OutputMsg",
		},
		{
			name:             "Process SSH connect message with OS terminal",
			embeddedTerminal: false,
			expectedMsgName:  "execMsg",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := MockAppState()
			state.EmbeddedTerminalEnabled = tt.embeddedTerminal
			model := New(context.TODO(), testutils.NewMockStorage(false), state, &mocklogger.Logger{})
			host := hostModel.Host{
				Address:       "localhost",
				SSHHostConfig: &sshconfig.Config{Hostname: "localhost"},
			}
			cmd := model.dispatchProcessSSHConnect(message.RunProcessSSHConnect{Host: host})
			require.Equal(t, tt.expectedMsgName, reflect.TypeOf(cmd()).Name())
		})
	}
}

func TestDispatchProcess_dispatchProcessSSHLoadConfig(t *testing.T) {
	model := New(context.TODO(), testutils.NewMockStorage(true), MockAppState(), &mocklogger.Logger{})
	host := hostModel.Host{Address: "localhost"}
	cmd := model.dispatchProcessSSHLoadConfig(message.RunProcessSSHLoadConfig{Host: host})
	require.IsType(t, message.RunProcessSuccess{}, cmd())
}

func TestDispatchProcess_Foreground(t *testing.T) {
	// Create a model
	model := New(context.TODO(), testutils.NewMockStorage(true), MockAppState(), &mocklogger.Logger{})

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

func TestDispatchProcess_Background_OK(t *testing.T) {
	// Create a model
	model := New(context.TODO(), testutils.NewMockStorage(true), MockAppState(), &mocklogger.Logger{})
	cmd := lo.Ternary(runtime.GOOS == "windows", "cmd /C echo test", "echo test")
	// Test case: Successful process execution
	validProcess := utils.BuildProcess(cmd)
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
	model := New(context.TODO(), testutils.NewMockStorage(true), MockAppState(), &mocklogger.Logger{})

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
	model := New(context.TODO(), testutils.NewMockStorage(true), MockAppState(), &mocklogger.Logger{})
	given := message.RunProcessSuccess{
		ProcessType: constant.ProcessTypeSSHLoadConfig,
		StdOut:      "hostname localhost\r\nport 2222\r\nidentityfile /tmp\r\nuser root",
	}

	expected := message.HostSSHConfigLoadComplete{
		HostID: 0,
		Config: sshconfig.Config{
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
			model := New(context.TODO(), testutils.NewMockStorage(true), MockAppState(), &mocklogger.Logger{})
			model.handleProcessSuccess(tt.msg)
			require.Equal(t, tt.expected.modelMessage, model.viewMessageContent)
			require.Equal(t, (int)(model.appState.CurrentView), tt.expected.viewState)
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
			model := New(context.TODO(), testutils.NewMockStorage(true), MockAppState(), &mocklogger.Logger{})
			model.handleProcessError(tt.msg)
			require.Equal(t, tt.expected.modelMessage, model.viewMessageContent)
			require.Equal(t, (int)(model.appState.CurrentView), tt.expected.viewState)
		})
	}
}

func TestDispatchProcessSSHCopyID(t *testing.T) {
	model := New(context.TODO(), testutils.NewMockStorage(true), MockAppState(), &mocklogger.Logger{})

	// Prepare a host with SSH config
	hostWithConfig := hostModel.Host{}
	hostWithConfig.SSHHostConfig = &sshconfig.Config{
		IdentityFile: "/tmp/id_rsa",
		Hostname:     "localhost",
	}

	msg := message.RunProcessSSHCopyID{
		Host: hostWithConfig,
	}

	cmd := model.dispatchProcessSSHCopyID(msg)
	// Check that cmd returns a tea.execMsg (unexported type, so check type name)
	require.Equal(t, "execMsg", reflect.TypeOf(cmd()).Name())
}

func TestUpdate_RunProcessErrorOccurred(t *testing.T) {
	model := New(context.TODO(), testutils.NewMockStorage(false), MockAppState(), &mocklogger.Logger{})
	msg := message.RunProcessErrorOccurred{
		ProcessType: constant.ProcessTypeSSHCopyID,
		StdOut:      "mock out message",
		StdErr:      "mock error message",
	}

	m, _ := model.Update(msg)
	require.Equal(t, "mock error message\nDetails: mock out message", m.(*MainModel).viewMessageContent)
	require.Equal(t, state.ViewMessage, m.(*MainModel).appState.CurrentView)
}

func TestUpdate_ExitWithError(t *testing.T) {
	model := New(context.TODO(), testutils.NewMockStorage(false), MockAppState(), &mocklogger.Logger{})
	msg := message.ExitWithError{
		Err: errors.New("mock error message"),
	}

	m, cmd := model.Update(msg)

	require.Equal(t, cmd(), tea.Quit())
	require.Equal(t, "mock error message", m.(*MainModel).exitError.Error())
}

func Test_handleKeyEvent(t *testing.T) {
	model := New(context.TODO(), testutils.NewMockStorage(false), MockAppState(), &mocklogger.Logger{})
	model.viewMessageContent = "some output of the process which just has finished"
	model.appState.CurrentView = state.ViewMessage
	model.Update(tea.KeyPressMsg{})

	// When user is facing a message from the process which has just finished and presses any key,
	// the view should switch to the host list and the message content should be cleared.
	require.Equal(t, state.ViewHostList, model.appState.CurrentView)
	require.Empty(t, model.viewMessageContent)
}

func Test_handleKeyEvent_CtrlC(t *testing.T) {
	// Check that CTRL+C is not passed to bubbletea when we're in SSH session.
	// This signal should be processed by the SSH session itself.
	tests := []struct {
		name        string
		currentView state.View
		sessionView tea.Model
		wantCmd     tea.Cmd
	}{{
		name:        "Ctrl+C should be ignored when current view is ViewSSHSession",
		currentView: state.ViewSSHSession,
		wantCmd:     nil,
	}, {
		name:        "Ctrl+C should quit the application when current view is not ViewSSHSession",
		currentView: state.ViewHostList,
		wantCmd:     tea.Quit,
	}}

	ctrlC := tea.KeyPressMsg{
		Mod:  uv.ModCtrl,
		Code: 'c',
	}

	var childModel tea.Model = modelFunc{
		init:   func() tea.Cmd { return nil },
		update: func(_ tea.Msg) (tea.Model, tea.Cmd) { return nil, nil },
		view:   func() tea.View { return tea.NewView("") },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := New(context.TODO(), testutils.NewMockStorage(false), MockAppState(), &mocklogger.Logger{})
			model.appState.CurrentView = tt.currentView
			model.modelSSHSession = childModel
			_, cmd := model.Update(ctrlC)
			if tt.wantCmd == nil {
				require.Nil(t, cmd)
			} else {
				require.Equal(t, tt.wantCmd(), cmd())
			}
		})
	}
}

func Test_view(t *testing.T) {
	fakeModelFactory := func(modelViewContent string) tea.Model {
		return modelFunc{
			init:   func() tea.Cmd { return nil },
			update: func(_ tea.Msg) (tea.Model, tea.Cmd) { return nil, nil },
			view: func() tea.View {
				return tea.NewView(modelViewContent)
			},
		}
	}

	m := New(context.TODO(), testutils.NewMockStorage(false), MockAppState(), &mocklogger.Logger{})
	// There will be no output without setting proper size of the viewport.
	m.viewport = viewport.New(viewport.WithHeight(1))
	m.modelGroupList = fakeModelFactory("mock group list")
	m.modelHostList = fakeModelFactory("mock host list")
	m.viewMessageContent = "mock message content"
	m.modelHostEdit = fakeModelFactory("mock host edit")
	m.modelSSHSession = fakeModelFactory("mock ssh session")

	tests := []struct {
		name     string
		appState state.View
		expected string
	}{
		{
			name:     "View should return group list when app state is ViewGroupList",
			appState: state.ViewGroupList,
			expected: "mock group list",
		},
		{
			name:     "View should return host list when app state is ViewHostList",
			appState: state.ViewHostList,
			expected: "mock host list",
		},
		{
			name:     "View should return message content when app state is ViewMessage",
			appState: state.ViewMessage,
			expected: "mock message content",
		},
		{
			name:     "View should return host edit when app state is ViewEditItem",
			appState: state.ViewEditItem,
			expected: "mock host edit",
		},
		{
			name:     "View should return ssh session when app state is ViewSSHSession",
			appState: state.ViewSSHSession,
			expected: "mock ssh session",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m.appState.CurrentView = tt.appState
			m.viewport.SetWidth(len(tt.expected))
			require.Equal(t, tt.expected, m.View().Content)
		})
	}
}

// ---------------------------------

func MockAppState() *state.State {
	return &state.State{}
}

type modelFunc struct {
	init   func() tea.Cmd
	update func(tea.Msg) (tea.Model, tea.Cmd)
	view   func() tea.View
}

func (m modelFunc) Init() tea.Cmd                           { return m.init() }
func (m modelFunc) Update(msg tea.Msg) (tea.Model, tea.Cmd) { return m.update(msg) }
func (m modelFunc) View() tea.View                          { return m.view() }
