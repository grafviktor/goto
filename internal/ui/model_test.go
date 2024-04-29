package ui

import (
	"context"
	"os"
	"reflect"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/grafviktor/goto/internal/state"
	"github.com/grafviktor/goto/internal/test"
	"github.com/grafviktor/goto/internal/ui/message"
	"github.com/grafviktor/goto/internal/utils"
)

func TestNew(t *testing.T) {
	model := New(context.TODO(), test.NewMockStorage(true), MockAppState(), &test.MockLogger{})
	require.NotNil(t, model)
}

func TestUpdate_KeyMsg(t *testing.T) {
	// Random key test - make sure that the app reacts on Ctrl+C
	model := New(context.TODO(), test.NewMockStorage(true), MockAppState(), &test.MockLogger{})
	_, cmd := model.Update(tea.KeyMsg{
		Type: tea.KeyCtrlC,
	})

	assert.NotNil(t, model)
	require.IsType(t, tea.QuitMsg{}, cmd(), "Wrong message type")
}

func TestUpdate_TerminalSizePolling(t *testing.T) {
	// Ensure that when the model receives TerminalSizePolling it autogenerates 'WindowSizeMsg'
	model := New(context.TODO(), test.NewMockStorage(true), MockAppState(), &test.MockLogger{})
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

func TestDispatchProcess_Foreground(t *testing.T) {
	// Create a model
	model := New(context.TODO(), test.NewMockStorage(true), MockAppState(), &test.MockLogger{})

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
	model := New(context.TODO(), test.NewMockStorage(true), MockAppState(), &test.MockLogger{})

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
	model := New(context.TODO(), test.NewMockStorage(true), MockAppState(), &test.MockLogger{})

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

// ---------------------------------

func MockAppState() *state.ApplicationState {
	return &state.ApplicationState{}
}
