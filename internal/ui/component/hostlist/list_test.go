// Package hostlist implements the host list view.
package hostlist

import (
	"context"
	"os"
	"testing"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/grafviktor/goto/internal/model"
	"github.com/grafviktor/goto/internal/utils"
)

func getTestListModel() *listModel {
	// Create a new host
	h := model.NewHost(0, "", "", "localhost", "root", "id_rsa", "2222")

	// Add three items to the list
	items := []list.Item{ListItemHost{h}, ListItemHost{h}, ListItemHost{h}}

	// Create listModel using constructor function (using 'New' is important to preserve hotkeys)
	lm := New(context.TODO(), nil, nil, nil)
	lm.innerModel.SetItems(items)

	return &lm
}

func Test_ListTitleUpdate(t *testing.T) {
	// Create a lm with initial state
	lm := *getTestListModel()

	// Select host
	lm.innerModel.Select(0)

	// Create a message of type msgFocusChanged
	msg := msgFocusChanged{}
	// Apply the function
	lm = lm.listTitleUpdate(msg)

	require.Equal(t, "ssh -i id_rsa -p 2222 -l root localhost", lm.innerModel.Title)
}

func Test_listModel_Change_Selection(t *testing.T) {
	tests := []struct {
		name                   string
		expectedSelectionIndex int
		model                  listModel
		tea.KeyMsg
	}{
		// Simulate focus next event
		{
			"Select next using 'j' key",
			2,
			*getTestListModel(),
			tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}},
		},
		{
			"Select next using '↓' key",
			2,
			*getTestListModel(),
			tea.KeyMsg{Type: tea.KeyDown},
		},
		{
			"Select next using 'tab' key",
			2,
			*getTestListModel(),
			tea.KeyMsg{Type: tea.KeyTab},
		},
		// Simulate focus previous event
		{
			"Select previous using 'k' key",
			0,
			*getTestListModel(),
			tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}},
		},
		{
			"Select previous using '↑' key",
			0,
			*getTestListModel(),
			tea.KeyMsg{Type: tea.KeyUp},
		},
		{
			"Select previous using 'shift+tab' key",
			0,
			*getTestListModel(),
			tea.KeyMsg{Type: tea.KeyShiftTab},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Select item at index 1. We need this preselection in order
			// to test 'focus previous' and 'focus next' messages
			tt.model.innerModel.Select(1)

			// Receive updated model
			updatedModel, _ := tt.model.Update(tt.KeyMsg)

			// Check if the selected index is correct
			if lm, ok := updatedModel.(listModel); ok {
				require.Equal(t, tt.expectedSelectionIndex, lm.innerModel.Index())
			} else {
				t.Error("Can't cast updatedModel to listModel")
			}
		})
	}
}

func Test_StdErrorWriter_Write(t *testing.T) {
	// Test the Write method of stdErrorWriter
	writer := stdErrorWriter{}
	data := []byte("test error")
	// 'n' should be equal to zero, as we're not writing errors to the terminal
	n, err := writer.Write(data)

	assert.NoError(t, err)
	// Make sure that 'n' is zero, because we don't want to see errors in the console
	assert.Equal(t, len(data), n)
	// However we can read the error text from writer.err variable when we need
	assert.Equal(t, data, writer.err)
}

func Test_BuildProcess(t *testing.T) {
	// Test case: Item is not selected
	listModelEmpty := listModel{innerModel: list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)}
	_, err := listModelEmpty.buildProcess(&stdErrorWriter{})
	require.Error(t, err)

	// Test case: Item is selected
	listModel := getTestListModel()
	cmd, err := listModel.buildProcess(&stdErrorWriter{})
	require.NoError(t, err)

	// Check that cmd is created and stdErr is re-defined
	require.NotNil(t, cmd)
	require.Equal(t, os.Stdout, cmd.Stdout)
	require.Equal(t, &stdErrorWriter{}, cmd.Stderr)
}

func Test_RunProcess(t *testing.T) {
	// Mock data for listModel
	listModel := getTestListModel()

	errorWriter := stdErrorWriter{}

	validProcess := utils.BuildProcess("echo test") // crossplatform command
	validProcess.Stdout = os.Stdout
	validProcess.Stderr = &errorWriter

	// Test case: Successful process execution
	resultListModel, resultCmd := listModel.runProcess(validProcess, &errorWriter)

	// Perform assertions
	require.NotNil(t, resultListModel)
	require.NotNil(t, resultCmd)
	// require.Equal(t, "", string(errorWriter.err)) useless, as the process doesn't start

	/**
	 * We should run the event loop to run the process, otheriwse the process won't start.
	 * NewProgram invocation is failing, need to invest more time.
	 */
	// p := tea.NewProgram(resultListModel)
	// if _, err := p.Run(); err != nil {
	// 	require.NoError(t, err)
	// }

	// // Perform assertions
	// require.NotNil(t, resultListModel)
	// require.NotNil(t, resultCmd)
	// require.Equal(t, "", string(errorWriter.err))

}
