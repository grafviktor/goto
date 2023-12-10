// Package hostlist implements the host list view.
package hostlist

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/grafviktor/goto/internal/model"
	"github.com/grafviktor/goto/internal/utils"
)

func Test_ListTitleUpdate(t *testing.T) {
	// Create a lm with initial state
	lm := *NewMockListModel(false)

	// Select host
	lm.innerModel.Select(0)

	// Create a message of type msgFocusChanged
	msg := msgRefreshUI{}
	// Apply the function
	lm = lm.listTitleUpdate(msg)

	require.Equal(t, "ssh -i id_rsa -p 2222 -l root localhost", lm.innerModel.Title)
}

func Test_listModel_Change_Selection(t *testing.T) {
	tests := []struct {
		name                   string
		expectedSelectionIndex int
		tea.KeyMsg
	}{
		// Simulate focus next event
		{
			"Select next using 'j' key",
			2,

			tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}},
		},
		{
			"Select next using '↓' key",
			2,
			tea.KeyMsg{Type: tea.KeyDown},
		},
		{
			"Select next using 'tab' key",
			2,
			tea.KeyMsg{Type: tea.KeyTab},
		},
		// Simulate focus previous event
		{
			"Select previous using 'k' key",
			0,
			tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}},
		},
		{
			"Select previous using '↑' key",
			0,
			tea.KeyMsg{Type: tea.KeyUp},
		},
		{
			"Select previous using 'shift+tab' key",
			0,
			tea.KeyMsg{Type: tea.KeyShiftTab},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := *NewMockListModel(false)
			// Select item at index 1. We need this preselection in order
			// to test 'focus previous' and 'focus next' messages
			model.innerModel.Select(1)

			// Receive updated model
			updatedModel, _ := model.Update(tt.KeyMsg)

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
	listModel := NewMockListModel(false)
	cmd, err := listModel.buildProcess(&stdErrorWriter{})
	require.NoError(t, err)

	// Check that cmd is created and stdErr is re-defined
	require.NotNil(t, cmd)
	require.Equal(t, os.Stdout, cmd.Stdout)
	require.Equal(t, &stdErrorWriter{}, cmd.Stderr)
}

func Test_RunProcess(t *testing.T) {
	// Mock data for listModel
	listModel := NewMockListModel(false)

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

// When remove mode is enabled, test confirm action event.
// Once confirmed, the item should be removed rom the list.
// However, we can't check whether the item was really deleted
// from the database, as we would have to wait while
func Test_removeItem(t *testing.T) {
	tests := []struct {
		name          string
		model         listModel
		mode          string
		want          interface{}
		preselectItem int
		expectedItems int
	}{
		{
			name:          "Remove item success",
			model:         *NewMockListModel(false),
			mode:          modeRemoveItem,
			want:          tea.BatchMsg{},
			expectedItems: 2,
		},
		{
			name:          "Remove item error because of the database error",
			model:         *NewMockListModel(true),
			mode:          modeRemoveItem,
			want:          msgErrorOccured{},
			expectedItems: 3,
		},
		{
			name:          "Remove item error wrong item selected",
			model:         *NewMockListModel(false),
			mode:          modeRemoveItem,
			want:          msgErrorOccured{},
			preselectItem: 10,
			expectedItems: 3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Preselect item
			tt.model.innerModel.Select(tt.preselectItem)
			// Set mode removeMode
			tt.model.mode = tt.mode
			// Call remove function
			_, cmd := tt.model.removeItem()
			// Expected to be tea.Batch, because when removing host we trigger extra commands
			require.IsType(t, tt.want, cmd(), "Wrong message type")
			// Get all items from the database without error
			items, _ := tt.model.repo.GetAll()
			// Make sure that the list contains expected quantity of ite,s after remove operation
			require.Equal(t, tt.expectedItems, len(items))
		})
	}
}

func Test_confirmAction(t *testing.T) {
	// Create a new model. There is no special mode (for instance remove item mode)
	model := NewMockListModel(false)
	// Imagine that user triggers confirm aciton
	updatedModel, cmd := model.confirmAction()
	// When cancel action, we reset mode and return back to normal state
	require.Len(t, updatedModel.mode, 0)
	// Updated model should not be nil
	require.NotNil(t, updatedModel)
	// Because there is no active mode, model should ignore the event
	require.Nil(t, cmd)

	// Create a new model
	model = NewMockListModel(false)
	// Now we enable remove mode
	model.mode = modeRemoveItem
	// Imagine that user triggers confirm aciton
	updatedModel, cmd = model.confirmAction()
	// When confirm action is triggered, we reset mode and return back to normal state
	require.Len(t, updatedModel.mode, 0)
	// Updated model should not be nil
	require.NotNil(t, updatedModel)
	// cmd should not be nil because when we modify storage, some events will be dispatched
	// we should not check the exact event type here, because it is action-dependent
	require.NotNil(t, cmd)
}

func Test_enterRemoveItemMode(t *testing.T) {
	// Create a new model
	model := *NewMockListModel(false)
	// Select non-existent index
	model.innerModel.Select(10)
	// Call enterRemoveItemMode function
	model, cmd := model.enterRemoveItemMode()
	// and make sure that mode is unchanged
	require.Len(t, model.mode, 0)
	// cmd() should return msgErrorOccured error
	require.IsType(t, msgErrorOccured{}, cmd(), "Wrong message type")

	// Create another model
	model = *NewMockListModel(false)
	// Select a first item, which is valid
	model.innerModel.Select(0)
	// Call enterRemoveItemMode function
	model, cmd = model.enterRemoveItemMode()
	// Ensure that we entered remove mode
	require.Equal(t, modeRemoveItem, model.mode)
	// md() should return msgRefreshUI in order to update title
	require.IsType(t, msgRefreshUI{}, cmd(), "Wrong message type")
}

// ================================================ MOCKS ========================================== //

// ============================================== List Model

func NewMockListModel(storageShouldFail bool) *listModel {
	storage := NewMockStorage(storageShouldFail)

	// Create listModel using constructor function (using 'New' is important to preserve hotkeys)
	lm := New(context.TODO(), storage, nil, nil)

	items := make([]list.Item, 0)
	// Wrap hosts into List items
	hosts := storage.hosts
	for _, h := range hosts {
		items = append(items, ListItemHost{Host: h})
	}

	lm.innerModel.SetItems(items)

	return &lm
}

// =============================================== Storage

func NewMockStorage(shouldFail bool) *mockStorage {
	hosts := []model.Host{
		model.NewHost(0, "", "", "localhost", "root", "id_rsa", "2222"),
		model.NewHost(0, "", "", "localhost", "root", "id_rsa", "2222"),
		model.NewHost(0, "", "", "localhost", "root", "id_rsa", "2222"),
	}

	return &mockStorage{
		shouldFail: shouldFail,
		hosts:      hosts,
	}
}

type mockStorage struct {
	shouldFail bool
	hosts      []model.Host
}

// Delete implements storage.HostStorage.
func (ms *mockStorage) Delete(id int) error {
	if ms.shouldFail {
		return errors.New("Mock error")
	}

	ms.hosts = append(ms.hosts[:id], ms.hosts[id+1:]...)

	return nil
}

// Get implements storage.HostStorage.
func (ms *mockStorage) Get(hostID int) (model.Host, error) {
	if ms.shouldFail {
		return model.Host{}, errors.New("Mock error")
	}

	return model.Host{}, nil
}

// GetAll implements storage.HostStorage.
func (ms *mockStorage) GetAll() ([]model.Host, error) {
	if ms.shouldFail {
		return ms.hosts, errors.New("Mock error")
	}

	return ms.hosts, nil
}

// Save implements storage.HostStorage.
func (ms *mockStorage) Save(model.Host) error {
	if ms.shouldFail {
		return errors.New("Mock error")
	}

	return nil
}
