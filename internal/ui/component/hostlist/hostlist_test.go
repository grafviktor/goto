// Package hostlist implements the host list view.
package hostlist

import (
	"context"
	"testing"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/grafviktor/goto/internal/constant"
	"github.com/grafviktor/goto/internal/state"
	"github.com/grafviktor/goto/internal/test"
	"github.com/grafviktor/goto/internal/ui/message"
	"github.com/grafviktor/goto/internal/utils"
)

func Test_ListTitleUpdate(t *testing.T) {
	// Create a lm with initial state
	lm := *NewMockListModel(false)
	lm.logger = &test.MockLogger{}

	// Select host
	lm.Select(0)

	// Apply the function
	lm.listTitleUpdate()

	require.Equal(t, "ssh -i id_rsa -p 2222 -l root localhost", lm.Title)
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
			model.Select(1)

			// Receive updated model
			model.Update(tt.KeyMsg)

			// Check if the selected index is correct
			require.Equal(t, tt.expectedSelectionIndex, model.Index())
		})
	}
}

func Test_StdErrorWriter_Write(t *testing.T) {
	// Test the Write method of ProcessBufferWriter
	writer := utils.ProcessBufferWriter{}
	data := []byte("test error")
	n, err := writer.Write(data)

	assert.NoError(t, err)
	// Make sure that 'n' is equal to the data length which we sent to the writer
	assert.Equal(t, len(data), n)
	// However we can read the error text from writer.err variable when we need
	assert.Equal(t, data, writer.Output)
}

// When remove mode is enabled, test confirm action event.
// Once confirmed, the item should be removed rom the list.
// However, we can't check whether the item was really deleted
// from the database, as we would have to wait while
func TestRemoveItem(t *testing.T) {
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
			want:          MsgRefreshRepo{},
			expectedItems: 2,
		},
		{
			name:          "Remove item error because of the database error",
			model:         *NewMockListModel(true),
			mode:          modeRemoveItem,
			want:          msgErrorOccurred{},
			expectedItems: 3,
		},
		{
			name:          "Remove item error wrong item selected",
			model:         *NewMockListModel(false),
			mode:          modeRemoveItem,
			want:          msgErrorOccurred{},
			preselectItem: 10,
			expectedItems: 3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.model.logger = &test.MockLogger{}
			// Preselect item
			tt.model.Select(tt.preselectItem)
			// Set mode removeMode
			tt.model.mode = tt.mode
			// Call remove function
			cmd := tt.model.removeItem()
			// Expected to be tea.Batch, because when removing host we trigger extra commands
			require.IsType(t, tt.want, cmd(), "Wrong message type")
			// Get all items from the database without error
			items, _ := tt.model.repo.GetAll()
			// Make sure that the list contains expected quantity of ite,s after remove operation
			require.Equal(t, tt.expectedItems, len(items))
		})
	}
}

func TestConfirmAction(t *testing.T) {
	// Create a new model. There is no special mode (for instance remove item mode)
	model := NewMockListModel(false)
	model.logger = &test.MockLogger{}
	// Imagine that user triggers confirm aciton
	cmd := model.confirmAction()
	// When cancel action, we reset mode and return back to normal state
	require.Len(t, model.mode, 0)
	// Updated model should not be nil
	require.NotNil(t, model)
	// Because there is no active mode, model should ignore the event
	require.Nil(t, cmd)

	// Create a new model
	model = NewMockListModel(false)
	model.logger = &test.MockLogger{}
	// Now we enable remove mode
	model.mode = modeRemoveItem
	// Imagine that user triggers confirm aciton
	cmd = model.confirmAction()
	// When confirm action is triggered, we reset mode and return back to normal state
	require.Len(t, model.mode, 0)
	// Updated model should not be nil
	require.NotNil(t, model)
	// cmd should not be nil because when we modify storage, some events will be dispatched
	// we should not check the exact event type here, because it is action-dependent
	require.NotNil(t, cmd)
}

func TestEnterRemoveItemMode(t *testing.T) {
	// Create a new model
	model := *NewMockListModel(false)
	model.logger = &test.MockLogger{}
	// Select non-existent index
	model.Select(10)
	// Call enterRemoveItemMode function
	cmd := model.enterRemoveItemMode()
	// and make sure that mode is unchanged
	require.Len(t, model.mode, 0)
	// cmd() should return msgErrorOccurred error
	require.IsType(t, msgErrorOccurred{}, cmd(), "Wrong message type")

	// Create another model
	model = *NewMockListModel(false)
	model.logger = &test.MockLogger{}
	// Select a first item, which is valid
	model.Select(0)
	// Call enterRemoveItemMode function
	cmd = model.enterRemoveItemMode()
	// Ensure that we entered remove mode
	require.Equal(t, modeRemoveItem, model.mode)
	// cmd() should return msgRefreshUI in order to update title
	require.IsType(t, msgRefreshUI{}, cmd(), "Wrong message type")
}

func TestExitRemoveItemMode(t *testing.T) {
	// Create a new model
	model := *NewMockListModel(false)
	model.logger = &test.MockLogger{}
	// Select a first item, which is valid
	model.Select(0)
	// Call enterRemoveItemMode function
	model.enterRemoveItemMode()
	// Ensure that we entered remove mode
	require.Equal(t, modeRemoveItem, model.mode)

	// Reject the action by pressing 'n' (it can be any key apart from 'y')
	_, cmd := model.Update(tea.KeyMsg{
		Type:  -1, // Type '-1' should be equal to 'KeyRunes'
		Runes: []rune{'n'},
	})

	// Ensure that model exited remove move
	require.Equal(t, modeDefault, model.mode)

	// cmd() should return msgRefreshUI in order to update title
	require.IsType(t, msgRefreshUI{}, cmd(), "Wrong message type")
}

func TestListTitleUpdate(t *testing.T) {
	// 1 Call listTitleUpdate when host is not selected
	model := *NewMockListModel(false)
	model.logger = &test.MockLogger{}
	// Select non-existent item
	model.Select(10)
	// Call listTitleUpdate function, but it will fail, however without throwing any errors
	model.listTitleUpdate()
	// Check that model is not nil
	require.NotNil(t, model)

	// 2 Call listTitleUpdate when removeMode is active
	model = *NewMockListModel(false)
	model.logger = &test.MockLogger{}
	// Select a host by valid index
	model.Select(0)
	// Enter remove mode
	model.enterRemoveItemMode()
	// Call listTitleUpdate function
	model.listTitleUpdate()
	// Check that app is now asking for a confirmation before delete
	require.Equal(t, "delete \"Mock Host 1\" ? (y/N)", model.Title)

	// 3 Call listTitleUpdate selected a host
	model = *NewMockListModel(false)
	model.logger = &test.MockLogger{}
	// Select a host by valid index
	model.Select(0)
	// Call listTitleUpdate function
	model.listTitleUpdate()
	// Check that app is displaying ssh connection string
	require.Equal(t, "ssh -i id_rsa -p 2222 -l root localhost", model.Title)
}

func TestListModel_title_when_app_just_starts(t *testing.T) {
	// This is just a sanity test, which checks title update function
	model := *NewMockListModel(false)
	model.logger = &test.MockLogger{}
	// When app just starts, it should display "press 'n' to add a new host"
	require.Equal(t, "press 'n' to add a new host", model.Title)
	// When press 'down' key, it should display a proper ssh connection string
	model.Update(tea.KeyMsg{Type: tea.KeyDown})
	// Calling refresh UI manually, otherwise would have to put time.Sleep function
	model.Update(msgRefreshUI{})
	require.Equal(t, "ssh -i id_rsa -p 2222 -l root localhost", model.Title)
}

func TestListModel_title_when_filter_is_enabled(t *testing.T) {
	// Test bugfix for https://github.com/grafviktor/goto/issues/37
	model := *NewMockListModel(false)
	model.logger = &test.MockLogger{}
	assert.Equal(t, model.FilterState(), list.Unfiltered)
	// Enable filter
	model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	assert.Equal(t, model.FilterState(), list.Filtering)
	model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'3'}})
	// Press down key and make sure that title is properly updated
	model.Update(tea.KeyMsg{Type: tea.KeyDown})
	assert.Equal(t, model.FilterState(), list.FilterApplied)
	require.Equal(t, "ssh -i id_rsa -p 2222 -l root localhost", model.Title)
}

func TestListModel_refreshRepo(t *testing.T) {
	// Test refreshRepo function with normal storage behavior
	storageShouldFail := false
	storage := test.NewMockStorage(storageShouldFail)
	fakeAppState := state.ApplicationState{Selected: 1}
	lm := New(context.TODO(), storage, &fakeAppState, &test.MockLogger{})
	teaCmd := lm.refreshRepo(nil)

	var dst []tea.Msg
	test.CmdToMessage(teaCmd, &dst)
	require.Contains(t, dst, msgRefreshUI{})

	// Check that hosts are filtered by Title
	require.Equal(t, "Mock Host 1", lm.Items()[0].(ListItemHost).Title())
	require.Equal(t, "Mock Host 2", lm.Items()[1].(ListItemHost).Title())
	// Check that currently selected item is "1", as it is set in the fakeAppState object
	require.Equal(t, 1, lm.SelectedItem().(ListItemHost).ID)

	// Now test refreshRepo function simulating a broken storage
	storageShouldFail = true
	storage = test.NewMockStorage(storageShouldFail)
	lm = New(
		context.TODO(),
		storage,
		&state.ApplicationState{}, // we don't need app state, as error should be reported before we can even use it
		&test.MockLogger{},
	)
	lm.logger = &test.MockLogger{}
	teaCmd = lm.refreshRepo(nil)

	// Check that msgErrorOccurred{} was found among returned messages, which indicate that
	// something is wrong with the storage
	require.Equal(t, "mock error", teaCmd().(msgErrorOccurred).err.Error())
}

func TestListModel_editItem(t *testing.T) {
	// Test edit item function by making sure that it's returning correct messages

	// First case - when host is not selected in the list of hosts.
	// We should receive an error because there is nothing to edit
	storage := test.NewMockStorage(false)
	lm := New(
		context.TODO(),
		storage,
		&state.ApplicationState{}, // we don't need app state, as error should be reported before we can even use it
		&test.MockLogger{},
	)
	lm.logger = &test.MockLogger{}
	teaCmd := lm.editItem(nil)

	require.IsType(t, msgErrorOccurred{}, teaCmd())

	// Second case - we select a host from the list and sending a message to parent form
	// That a host with a certain ID is ready to be modified.
	//
	// Note, that here we use NewMockListModel instead of just 'list.New(...)' like in the first case
	// we need it to automatically preselect first item from the list of hosts and NewMockListModel
	// will do that for us
	lm = NewMockListModel(false)
	lm.logger = &test.MockLogger{}

	teaCmd = lm.editItem(nil)

	var dst []tea.Msg
	test.CmdToMessage(teaCmd, &dst)

	require.Contains(t, dst, OpenEditForm{HostID: 1})
	require.Contains(t, dst, message.RunProcessLoadSSHConfig{Host: lm.SelectedItem().(ListItemHost).Host})
}

func TestListModel_copyItem(t *testing.T) {
	// First case - test that we receive an error when item is not selected
	storageShouldFail := true
	storage := test.NewMockStorage(storageShouldFail)
	lm := New(context.TODO(), storage, &state.ApplicationState{}, &test.MockLogger{})
	teaCmd := lm.copyItem(nil)
	require.Equal(t, itemNotSelectedMessage, teaCmd().(msgErrorOccurred).err.Error())

	// Second case: storage is, OK and we have to ensure that copied host title as we expect it to be:
	lm = NewMockListModel(false)
	lm.logger = &test.MockLogger{}

	lm.copyItem(nil)
	host, err := lm.repo.Get(3)
	require.NoError(t, err)
	require.Equal(t, "Mock Host 1 (1)", host.Title)
}

func TestListModel_updateKeyMap(t *testing.T) {
	// Case 1: Test that if a host list contains items and item is selected, then all keyboard shortcuts are shown on the screen
	lm := *NewMockListModel(false)
	lm.logger = &test.MockLogger{}
	lm.Update(msgRefreshUI{})

	// Actually "displayedKeys" will also contain cursor up and cursor down and help keybindings,
	// but we're ignoring them in this test
	displayedKeys := lm.keyMap.ShortHelp()
	availableKeys := newDelegateKeyMap()

	require.Equal(t, 5, len(displayedKeys))
	require.Contains(t, displayedKeys, availableKeys.append)
	require.Contains(t, displayedKeys, availableKeys.clone)
	require.Contains(t, displayedKeys, availableKeys.connect)
	require.Contains(t, displayedKeys, availableKeys.edit)
	require.Contains(t, displayedKeys, availableKeys.remove)

	// Case 2: Test that if a host list contains items and item is NOT selected,
	// then some of the keyboard shortcuts should NOT be shown.

	// Now let's delete all elements from the database
	lm.repo.Delete(1)
	lm.repo.Delete(2)
	lm.repo.Delete(3)

	lm.Update(MsgRefreshRepo{})
	lm.Update(msgRefreshUI{})
	displayedKeys = lm.keyMap.ShortHelp()

	require.Equal(t, 1, len(displayedKeys))
	require.Contains(t, displayedKeys, availableKeys.append)
	require.NotContains(t, displayedKeys, availableKeys.clone)
	require.NotContains(t, displayedKeys, availableKeys.connect)
	require.NotContains(t, displayedKeys, availableKeys.edit)
	require.NotContains(t, displayedKeys, availableKeys.remove)
}

func TestUpdate_TeaSizeMsg(t *testing.T) {
	// Test that if model is ready, WindowSizeMsg message will update inner model size
	model := *NewMockListModel(false)
	model.logger = &test.MockLogger{}
	model.Update(tea.WindowSizeMsg{Width: 100, Height: 100})

	require.Greater(t, model.Height(), 0)
	require.Greater(t, model.Width(), 0)
}

func TestUpdate_SearchFunctionOfInnerModelIsNotRegressed(t *testing.T) {
	// Test that filtering is working properly

	// Create mock storage which contains hosts:
	// "Mock Host 1"
	// "Mock Host 2"
	// "Mock Host 3"
	storage := test.NewMockStorage(false)
	fakeAppState := state.ApplicationState{Selected: 1}

	// Create model
	model := New(context.TODO(), storage, &fakeAppState, &test.MockLogger{})
	model.logger = &test.MockLogger{}
	model.refreshRepo(nil)

	// Make sure there are 3 items in the collection
	assert.Len(t, model.VisibleItems(), 3)

	// Enable filtering mode
	model.Update(tea.KeyMsg{
		// KeyRunes equals to "-1". See
		// https://github.com/charmbracelet/bubbletea/blob/2ac3642f644d1c4ea67642910e77f7f56c58d2e9/key.go#L205
		Type:  tea.KeyRunes,
		Runes: []rune{'/'},
	})

	// Check that filtering mode is enabled
	assert.True(t, model.SettingFilter())

	// Now press "1" button. Only one item should left in the host list - with title: "Mock Host 1"
	_, cmds := model.Update(tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune{'1'},
	})

	// Extract batch messages from cmd
	msgs := []tea.Msg{}
	test.CmdToMessage(cmds, &msgs)

	// Feed all messages one by one to the model
	for _, msg := range msgs {
		model.Update(msg)
	}

	// Ensure, that only one item left in the list (which is "Mock Host 1")
	require.Len(t, model.VisibleItems(), 1)
}

func TestUpdate_ToggleBetweenNormalAndCompactLayout(t *testing.T) {
	// Create mock storage which contains hosts:
	// "Mock Host 1"
	// "Mock Host 2"
	// "Mock Host 3"
	storage := test.NewMockStorage(false)
	fakeAppState := state.ApplicationState{Selected: 1}

	// Create model
	model := New(context.TODO(), storage, &fakeAppState, &test.MockLogger{})
	model.logger = &test.MockLogger{}
	model.refreshRepo(nil)

	// Make sure there are 3 items in the collection
	assert.Len(t, model.VisibleItems(), 3)
	// Ensure that screen layout is not set
	layoutNotSet := constant.ScreenLayout("")
	assert.Equal(t, fakeAppState.ScreenLayout, layoutNotSet)

	// Toggle layout
	model.Update(tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune{'v'},
	})

	fakeAppState.ScreenLayout = constant.LayoutTight
	// Ensure that screen layout is equal to
	require.Equal(t, constant.LayoutTight, fakeAppState.ScreenLayout)

	// Toggle layout again and check that it's now set to "normal"
	model.Update(tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune{'v'},
	})

	require.Equal(t, constant.LayoutNormal, fakeAppState.ScreenLayout)
}

func TestBuildScreenLayout(t *testing.T) {
	layout := constant.LayoutNormal
	screenLayoutDelegate := NewHostDelegate(&layout, &test.MockLogger{})
	require.Equal(t, 1, screenLayoutDelegate.Spacing())
	require.True(t, screenLayoutDelegate.ShowDescription)

	layout = constant.LayoutTight
	screenLayoutDelegate = NewHostDelegate(&layout, &test.MockLogger{})
	require.Equal(t, 0, screenLayoutDelegate.Spacing())
	require.False(t, screenLayoutDelegate.ShowDescription)
}

func TestUpdate_HostFocusPreservedAfterClearFilterMessage(t *testing.T) {
	// Test that the same host is selected after we exit filter mode (clear filter with "Esc" button)

	// Create mock storage which contains hosts:
	// "Mock Host 1"
	// "Mock Host 2"
	// "Mock Host 3"
	storage := test.NewMockStorage(false)
	fakeAppState := state.ApplicationState{Selected: 1}

	// Create model
	model := New(context.TODO(), storage, &fakeAppState, &test.MockLogger{})
	model.logger = &test.MockLogger{}
	model.refreshRepo(nil)

	// Make sure there are 3 items in the collection
	assert.Len(t, model.VisibleItems(), 3)

	// Check that first item is selected
	assert.IsType(t, ListItemHost{}, model.SelectedItem())
	assert.Equal(t, "Mock Host 1", model.SelectedItem().(ListItemHost).Title())

	// Check that current status is "Unfiltered" and then enter filtering mode
	assert.Equal(t, list.Unfiltered, model.FilterState())
	model.Update(tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune{'/'},
	})
	assert.Equal(t, list.Filtering, model.FilterState())

	// When in filter mode type '2', so only "Mock Host 2" will become visible
	_, cmds := model.Update(tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune{'2'},
	})

	// Extract batch messages returned by the model
	msgs := []tea.Msg{}
	test.CmdToMessage(cmds, &msgs)

	// Send those messages back to the model
	for _, m := range msgs {
		model.Update(m)
	}

	assert.Len(t, model.VisibleItems(), 1)
}

// ==============================================
// ============================================== List Model
// ==============================================

func NewMockListModel(storageShouldFail bool) *listModel {
	storage := test.NewMockStorage(storageShouldFail)

	// Create listModel using constructor function (using 'New' is important to preserve hotkeys)
	lm := New(context.TODO(), storage, &state.ApplicationState{}, &test.MockLogger{})

	items := make([]list.Item, 0)
	// Wrap hosts into List items
	hosts := storage.Hosts
	for _, h := range hosts {
		items = append(items, ListItemHost{Host: h})
	}

	lm.SetItems(items)

	return lm
}
