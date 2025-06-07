package hostlist

import (
	"context"
	"errors"
	"testing"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/grafviktor/goto/internal/constant"
	"github.com/grafviktor/goto/internal/model/host"
	"github.com/grafviktor/goto/internal/model/sshconfig"
	"github.com/grafviktor/goto/internal/state"
	testutils "github.com/grafviktor/goto/internal/testutils"
	"github.com/grafviktor/goto/internal/testutils/mocklogger"
	"github.com/grafviktor/goto/internal/ui/message"
	"github.com/grafviktor/goto/internal/utils"
)

func TestListModel_Init(t *testing.T) {
	// Test Init function which loads data from storage
	storageShouldFail := false
	storage := testutils.NewMockStorage(storageShouldFail)
	fakeAppState := state.Application{Selected: 1}
	lm := New(context.TODO(), storage, &fakeAppState, &mocklogger.Logger{})
	teaCmd := lm.Init()

	var dst []tea.Msg
	testutils.CmdToMessage(teaCmd, &dst)
	require.Equal(t, dst, []tea.Msg{
		message.HostSelected{HostID: 1},
		message.RunProcessSSHLoadConfig{
			Host: host.Host{
				ID:               1,
				Title:            "Mock Host 1",
				Description:      "",
				Group:            "Group 1",
				Address:          "localhost",
				RemotePort:       "2222",
				LoginName:        "root",
				IdentityFilePath: "id_rsa",
				SSHHostConfig:    &sshconfig.Config{},
			},
		},
	})

	// Check that hosts are filtered by Title
	require.Equal(t, "Mock Host 1", lm.Items()[0].(ListItemHost).Title())
	require.Equal(t, "Mock Host 2", lm.Items()[1].(ListItemHost).Title())
	// Check that currently selected item is "1"
	require.Equal(t, 1, lm.SelectedItem().(ListItemHost).ID)

	// Test loadHosts function when a group is selected
	storage = testutils.NewMockStorage(false)
	lm = New(
		context.TODO(),
		storage,
		&state.Application{Group: "Group 2"},
		&mocklogger.Logger{},
	)
	lm.Init()
	require.Len(t, lm.Items(), 1)
	require.Equal(t, "Mock Host 2", lm.SelectedItem().(ListItemHost).Title())

	// Now test initialLoad function simulating a broken storage
	storageShouldFail = true
	storage = testutils.NewMockStorage(storageShouldFail)
	lm = New(
		context.TODO(),
		storage,
		&state.Application{}, // we don't need app state, as error should be reported before we can even use it
		&mocklogger.Logger{},
	)
	teaCmd = lm.Init()

	// Check that msgErrorOccurred{} was found among returned messages, which indicate that
	// something is wrong with the storage
	require.Equal(t, "mock error", teaCmd().(message.ErrorOccurred).Err.Error())
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

// When remove mode is enabled, test confirm action event.
// Once confirmed, the item should be removed rom the list.
// However, we can't check whether the item was really deleted
// from the database, as we would have to wait for a while
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
			preselectItem: 0, // It's already '0' by default. Just to be more explicit
			want: []tea.Msg{
				message.HideUINotification{ComponentName: "hostlist"},
				// Because we remote item "Mock Host 1" (which has index 0), we should ensure that next available item will be focused
				message.HostSelected{HostID: 2},
				message.RunProcessSSHLoadConfig{
					Host: host.Host{
						ID:               2,
						Title:            "Mock Host 2",
						Description:      "",
						Group:            "Group 2",
						Address:          "localhost",
						RemotePort:       "2222",
						LoginName:        "root",
						IdentityFilePath: "id_rsa",
						SSHHostConfig:    &sshconfig.Config{},
					},
				},
			},
			expectedItems: 2,
		},
		{
			// Checking this because there is a focusing problem in bubbles/list component
			// when remove last item, the component does not set focus on any item.
			name:          "Remove LAST item success",
			model:         *NewMockListModel(false),
			mode:          modeRemoveItem,
			preselectItem: 2, // We have 3 items in the mock storage. Selecting the last one
			want: []tea.Msg{
				message.HideUINotification{ComponentName: "hostlist"},
				// Because we remote item "Mock Host 1" (which has index 0), we should ensure that next available item will be focused
				message.HostSelected{HostID: 2},
				message.RunProcessSSHLoadConfig{
					Host: host.Host{
						ID:               2,
						Title:            "Mock Host 2",
						Description:      "",
						Group:            "Group 2",
						Address:          "localhost",
						RemotePort:       "2222",
						LoginName:        "root",
						IdentityFilePath: "id_rsa",
						SSHHostConfig:    &sshconfig.Config{},
					},
				},
			},
			expectedItems: 2,
		},
		{
			name:  "Remove item error because of the database error",
			model: *NewMockListModel(true),
			mode:  modeRemoveItem,
			want: []tea.Msg{
				message.ErrorOccurred{Err: errors.New("mock error")},
			},
			expectedItems: 3,
		},
		{
			name:  "Remove item error wrong item selected",
			model: *NewMockListModel(false),
			mode:  modeRemoveItem,
			want: []tea.Msg{
				message.ErrorOccurred{Err: errors.New("you must select an item")},
			},
			preselectItem: 10,
			expectedItems: 3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.model.logger = &mocklogger.Logger{}
			// Preselect item
			tt.model.Select(tt.preselectItem)
			// Set mode removeMode
			tt.model.mode = tt.mode
			// Call remove function
			cmd := tt.model.removeItem()
			// Cmd() can return a sequence or tea.Msg
			var actual []tea.Msg
			testutils.CmdToMessage(cmd, &actual)
			require.ElementsMatch(t, tt.want, actual, "Wrong message type")
			// Get all items from the database without error
			items, _ := tt.model.repo.GetAll()
			// Make sure that the list contains expected quantity of ite,s after remove operation
			require.Equal(t, tt.expectedItems, len(items))
		})
	}
}

func TestConfirmAction(t *testing.T) {
	// Test the fallback option when there is no active mode
	// Create a new model. There is no special mode (like for instance, "remove item" mode)
	model := NewMockListModel(false)
	model.logger = &mocklogger.Logger{}
	// Imagine that user triggers confirm action
	cmd := model.confirmAction()
	// When cancel action, we reset mode and return back to normal state
	require.Equal(t, model.mode, modeDefault)
	// Because there is no active mode, model should ignore the event
	require.Nil(t, cmd)

	// Now test remove item mode
	model = NewMockListModel(false)
	// Now we enable remove mode
	model.mode = modeRemoveItem
	// Imagine that user triggers confirm action
	cmd = model.confirmAction()
	// When confirm action is triggered, we reset mode and return back to normal state
	require.Equal(t, model.mode, modeDefault)
	// cmd should not be nil because when we modify storage, some Cmds will be dispatched
	require.IsType(t, tea.Cmd(nil), cmd)

	// Now test copy ssh ID item mode
	model = NewMockListModel(false)
	// Now we enable copy SSG ID mode
	model.mode = modeSSHCopyID
	// Imagine that user triggers confirm action
	cmd = model.confirmAction()
	// When confirm action is triggered, we reset mode and return back to normal state
	require.Equal(t, model.mode, modeDefault)
	// cmd should not be nil because when we modify storage, some Cmds will be dispatched
	require.IsType(t, tea.Cmd(nil), cmd)

	// Now test close app mode
	model = NewMockListModel(false)
	// Now we enable close application mode
	model.mode = modeCloseApp
	// Imagine that user triggers confirm action
	cmd = model.confirmAction()
	// When confirm action is triggered, we reset mode and return back to normal state
	require.Equal(t, model.mode, modeDefault)
	// If this mode is confirm the model should dispatch QuitMsg
	require.IsType(t, tea.QuitMsg{}, cmd())
}

func TestEnterSSHCopyIDMode(t *testing.T) {
	// Create a new model
	model := *NewMockListModel(false)
	// Select non-existent index
	model.Select(10)
	// Call enterRemoveItemMode function
	cmd := model.enterSSHCopyIDMode()
	// and make sure that mode is unchanged
	require.Len(t, model.mode, 0)
	// cmd() should return msgErrorOccurred error
	require.IsType(t, message.ErrorOccurred{}, cmd(), "Wrong message type")

	// Now select an existing item in the host list
	model.Select(0)
	// Call enterRemoveItemMode function
	cmd = model.enterSSHCopyIDMode()
	// cmd should be equal to nil
	require.Nil(t, cmd, "Wrong message type")
	// Ensure that we entered remove mode and title is updated
	require.Equal(t, modeSSHCopyID, model.mode)
	require.Equal(t, "copy ssh key to the remote host? (y/N)", utils.StripStyles(model.Title))
}

func TestEnterRemoveItemMode(t *testing.T) {
	// Create a new model
	model := *NewMockListModel(false)
	model.logger = &mocklogger.Logger{}
	// Select non-existent index
	model.Select(10)
	// Call enterRemoveItemMode function
	cmd := model.enterRemoveItemMode()
	// and make sure that mode is unchanged
	require.Len(t, model.mode, 0)
	// cmd() should return msgErrorOccurred error
	require.IsType(t, message.ErrorOccurred{}, cmd(), "Wrong message type")

	// Create another model
	model = *NewMockListModel(false)
	model.logger = &mocklogger.Logger{}
	// Select a first item, which is valid
	model.Select(0)
	// Call enterRemoveItemMode function
	cmd = model.enterRemoveItemMode()
	// cmd should be equal to nil
	require.Nil(t, cmd, "Wrong message type")
	// Ensure that we entered remove mode and title is updated
	require.Equal(t, modeRemoveItem, model.mode)
	require.Equal(t, "delete \"Mock Host 1\"? (y/N)", utils.StripStyles(model.Title))
}

func TestExitRemoveItemMode(t *testing.T) {
	// Create a new model
	model := *NewMockListModel(false)
	model.logger = &mocklogger.Logger{}
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

	expected := []tea.Msg{
		// Because we remote item "Mock Host 1" (which has index 0), we should ensure that next available item will be focused
		message.HostSelected{HostID: 1},
		message.RunProcessSSHLoadConfig{
			Host: host.Host{
				ID:               1,
				Title:            "Mock Host 1",
				Description:      "",
				Group:            "Group 1",
				Address:          "localhost",
				RemotePort:       "2222",
				LoginName:        "root",
				IdentityFilePath: "id_rsa",
				SSHHostConfig:    &sshconfig.Config{},
			},
		},
	}

	var actual []tea.Msg
	testutils.CmdToMessage(cmd, &actual)
	require.Equal(t, expected, actual, "Wrong message type")

	// Ensure that model exited remove move
	require.Equal(t, modeDefault, model.mode)
}

func TestListTitleUpdate(t *testing.T) {
	// 1 Call updateTitle when host is not selected
	model := *NewMockListModel(false)
	model.logger = &mocklogger.Logger{}
	// Select non-existent item
	model.Select(10)
	// Call updateTitle function, but it will fail, however without throwing any errors
	model.updateTitle()
	// Check that model is not nil
	require.NotNil(t, model)

	// 2 Call updateTitle when removeMode is active
	model = *NewMockListModel(false)
	model.logger = &mocklogger.Logger{}
	// Select a host by valid index
	model.Select(0)
	// Enter remove mode
	model.enterRemoveItemMode()
	// Call updateTitle function
	model.updateTitle()
	// Check that app is now asking for a confirmation before delete
	require.Equal(t, "delete \"Mock Host 1\"? (y/N)", utils.StripStyles(model.Title))

	// 3 Call updateTitle selected a host
	model = *NewMockListModel(false)
	model.logger = &mocklogger.Logger{}
	// Select a host by valid index
	model.Select(0)
	// Call updateTitle function
	model.updateTitle()
	// Check that app is displaying ssh connection string.
	require.Equal(t, "ssh -i id_rsa -p 2222 -l root localhost", utils.StripStyles(model.Title))

	// 4 Call updateTitle selected a host and group is selected.
	model = *NewMockListModel(false)
	model.logger = &mocklogger.Logger{}
	//
	model.appState.Group = "Group 1"
	// Select a host by valid index
	model.Select(0)
	// Call updateTitle function
	model.updateTitle()
	// Check that app is displaying ssh connection string prepended by a group abbreviation.
	require.Equal(t, "G1  ssh -i id_rsa -p 2222 -l root localhost", utils.StripStyles(model.Title))

	// 5 Call updateTitle when exiting app.
	model = *NewMockListModel(false)
	model.logger = &mocklogger.Logger{}
	// Enter exit app mode
	model.enterCloseAppMode()
	// Call updateTitle function
	model.updateTitle()
	// Check that app is asking the user for confirmation.
	require.Equal(t, "close app? (y/N)", utils.StripStyles(model.Title))
}

func TestListModel_title_when_app_just_starts(t *testing.T) {
	// This is just a sanity test, which checks title update function
	model := *NewMockListModel(false)
	model.logger = &mocklogger.Logger{}
	// When app just starts, it should display "press 'n' to add a new host"
	require.Equal(t, "press 'n' to add a new host", utils.StripStyles(model.Title))
	// When press 'down' key, it should display a proper ssh connection string
	model.Update(tea.KeyMsg{Type: tea.KeyDown})
	require.Equal(t, "ssh -i id_rsa -p 2222 -l root localhost", utils.StripStyles(model.Title))
}

func TestListModel_title_when_filter_is_enabled(t *testing.T) {
	// Test bugfix for https://github.com/grafviktor/goto/issues/37
	model := *NewMockListModel(false)
	model.logger = &mocklogger.Logger{}
	assert.Equal(t, model.FilterState(), list.Unfiltered)
	// Enable filter
	model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	assert.Equal(t, model.FilterState(), list.Filtering)
	model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'3'}})
	// Press down key and make sure that title is properly updated
	model.Update(tea.KeyMsg{Type: tea.KeyDown})
	assert.Equal(t, model.FilterState(), list.FilterApplied)
	require.Equal(t, "ssh -i id_rsa -p 2222 -l root localhost", utils.StripStyles(model.Title))
}

func TestListModel_editItem(t *testing.T) {
	// Test edit item function by making sure that it's returning correct messages

	// First case - when host is not selected in the list of hosts.
	// We should receive an error because there is nothing to edit
	storage := testutils.NewMockStorage(false)
	lm := New(
		context.TODO(),
		storage,
		&state.Application{}, // we don't need app state, as error should be reported before we can even use it
		&mocklogger.Logger{},
	)
	lm.logger = &mocklogger.Logger{}
	teaCmd := lm.editItem()

	require.IsType(t, message.ErrorOccurred{}, teaCmd())

	// Second case - we select a host from the list and sending a message to parent form
	// That a host with a certain ID is ready to be modified.
	//
	// Note, that here we use NewMockListModel instead of just 'list.New(...)' like in the first case
	// we need it to automatically preselect first item from the list of hosts and NewMockListModel
	// will do that for us
	lm = NewMockListModel(false)
	lm.logger = &mocklogger.Logger{}

	teaCmd = lm.editItem()

	var dst []tea.Msg
	testutils.CmdToMessage(teaCmd, &dst)

	require.Contains(t, dst, message.OpenViewHostEdit{HostID: 1})
	require.Contains(t, dst, message.RunProcessSSHLoadConfig{Host: lm.SelectedItem().(ListItemHost).Host})
}

func TestListModel_copyItem(t *testing.T) {
	// First case - test that we receive an error when item is not selected
	storageShouldFail := true
	storage := testutils.NewMockStorage(storageShouldFail)
	lm := New(context.TODO(), storage, &state.Application{}, &mocklogger.Logger{})
	teaCmd := lm.copyItem()
	require.Equal(t, itemNotSelectedErrMsg, teaCmd().(message.ErrorOccurred).Err.Error())

	// Second case: storage is OK, and we have to ensure that copied host title as we expect it to be:
	lm = NewMockListModel(false)
	lm.logger = &mocklogger.Logger{}

	lm.copyItem()
	host, err := lm.repo.Get(3)
	require.NoError(t, err)
	require.Equal(t, "Mock Host 1 (1)", host.Title)
}

func TestListModel_updateKeyMap(t *testing.T) {
	// Case 1: Test that if a host list contains items and item is selected, then all keyboard shortcuts are shown on the screen
	lm := *NewMockListModel(false)
	lm.logger = &mocklogger.Logger{}
	lm.Init()

	// Actually "displayedKeys" will also contain cursor up and cursor down and help keybindings,
	// but we're ignoring them in this test
	displayedKeys := lm.keyMap.ShortHelp()
	availableKeys := newDelegateKeyMap()

	require.Equal(t, 6, len(displayedKeys))
	require.Contains(t, displayedKeys, availableKeys.append)
	require.Contains(t, displayedKeys, availableKeys.clone)
	require.Contains(t, displayedKeys, availableKeys.connect)
	require.Contains(t, displayedKeys, availableKeys.edit)
	require.Contains(t, displayedKeys, availableKeys.remove)

	// Case 2: Test that if a host list does not contain any items,
	// then some of the keyboard shortcuts should NOT be shown.
	// Removing all hosts.
	lm.enterRemoveItemMode()
	lm.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	lm.enterRemoveItemMode()
	lm.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	lm.enterRemoveItemMode()
	lm.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})

	displayedKeys = lm.keyMap.ShortHelp()

	// Only "new" and "change group" shortcuts are available
	require.Equal(t, 2, len(displayedKeys))
	require.Contains(t, displayedKeys, availableKeys.append)
	require.NotContains(t, displayedKeys, availableKeys.clone)
	require.NotContains(t, displayedKeys, availableKeys.connect)
	require.NotContains(t, displayedKeys, availableKeys.edit)
	require.NotContains(t, displayedKeys, availableKeys.remove)
}

func TestUpdate_TeaSizeMsg(t *testing.T) {
	// Test that if model is ready, WindowSizeMsg message will update inner model size
	model := *NewMockListModel(false)
	model.logger = &mocklogger.Logger{}
	model.Update(tea.WindowSizeMsg{Width: 100, Height: 100})

	require.Greater(t, model.Height(), 0)
	require.Greater(t, model.Width(), 0)
}

func TestUpdate_HostSSHConfigLoaded(t *testing.T) {
	// Test that host receives SSH expectedConfig once HostConfigLoaded message is dispatched
	lm := *NewMockListModel(false)
	lm.Init()
	expectedConfig := sshconfig.Config{
		Hostname:     "mock_hostname",
		IdentityFile: "/tmp",
		Port:         "9999",
		User:         "mock_username",
	}
	lm.Update(message.HostSSHConfigLoaded{
		HostID: 1,
		Config: expectedConfig,
	})

	actualConfig := lm.Items()[0].(ListItemHost).SSHHostConfig
	require.Equal(t, &expectedConfig, actualConfig)
}

func TestUpdate_HostUpdated(t *testing.T) {
	// Test that host is updated when hostlist model receives message.HostUpdated.
	lm := *NewMockListModel(false)
	lm.Init()

	// Check that the host we're going to update exists and has the expected title
	require.Equal(t, lm.Items()[0].(ListItemHost).Title(), "Mock Host 1")

	updatedHost := host.Host{
		ID:               1,
		Title:            "Mock Host 11",
		Description:      "Mock Host Updated",
		Address:          "mock_hostname",
		RemotePort:       "9999",
		LoginName:        "mock_username",
		IdentityFilePath: "/tmp",
		SSHHostConfig:    nil,
	}

	lm.Update(message.HostUpdated{Host: updatedHost})
	require.Equal(t, updatedHost, lm.Items()[0].(ListItemHost).Host)

	// Also check that host is inserted into a correct position of the hostlist model
	updatedHost = host.Host{
		ID:               1,
		Title:            "zzz", // Title is now updated, the host should be positioned at the last index
		Description:      "Mock Host Updated",
		Address:          "mock_hostname",
		RemotePort:       "9999",
		LoginName:        "mock_username",
		IdentityFilePath: "/tmp",
		SSHHostConfig:    nil,
	}

	lm.Update(message.HostUpdated{Host: updatedHost})
	lastIndex := 2
	require.Equal(t, updatedHost, lm.Items()[lastIndex].(ListItemHost).Host)
}

func TestUpdate_HostCreated(t *testing.T) {
	// Test that when host is created it is appended to the host list and
	// its visual position in the list of hosts is correct
	lm := *NewMockListModel(false)
	lm.Init()

	require.Equal(t, lm.Items()[0].(ListItemHost).Title(), "Mock Host 1")

	createdHost1 := host.Host{
		ID:               999,
		Title:            "AAA new host", // Should be positioned first
		Description:      "Mock Host Updated",
		Address:          "mock_hostname",
		RemotePort:       "9999",
		LoginName:        "mock_username",
		IdentityFilePath: "/tmp",
		SSHHostConfig:    nil,
	}

	lm.Update(message.HostCreated{Host: createdHost1})
	require.Len(t, lm.Items(), 4, "Wrong host list size")
	require.Equal(t, createdHost1, lm.Items()[0].(ListItemHost).Host)

	// Also check that host is inserted into a correct position of the hostlist model
	createdHost2 := host.Host{
		ID:               666,
		Title:            "ZZZ new host", // Should be positioned at last index
		Description:      "Mock Host Updated",
		Address:          "mock_hostname",
		RemotePort:       "9999",
		LoginName:        "mock_username",
		IdentityFilePath: "/tmp",
		SSHHostConfig:    nil,
	}

	lm.Update(message.HostCreated{Host: createdHost2})
	require.Len(t, lm.Items(), 5, "Wrong host list size")
	lastIndex := 4 // because we have 5 hosts in total
	require.Equal(t, createdHost2, lm.Items()[lastIndex].(ListItemHost).Host)
}

func TestUpdate_GroupListSelectItem(t *testing.T) {
	// Test when select group, the model must reload hosts and reset filter
	model := New(
		context.TODO(),
		testutils.NewMockStorage(false),
		&state.Application{},
		&mocklogger.Logger{},
	)
	model.loadHosts()
	// Load All items from the collection
	require.Len(t, model.Items(), 3)
	// Enable filter
	model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	require.Equal(t, list.Filtering, model.FilterState())

	// Dispatch message
	model.Update(message.GroupSelected{Name: "Group 1"})

	// Now test, that there is only a single list in the collection
	require.Len(t, model.Items(), 1)
	// Filtering is off
	require.Equal(t, list.Unfiltered, model.FilterState())
}

func TestUpdate_msgHideNotification(t *testing.T) {
	// Test that title resets back to normal when hiding notification
	model := New(
		context.TODO(),
		testutils.NewMockStorage(false),
		&state.Application{},
		&mocklogger.Logger{},
	)
	model.loadHosts()
	model.Title = "Mock notification message"

	model.Update(message.HideUINotification{ComponentName: "hostlist"})

	// Ensure that notification returned back to normal when hid the notification message.
	require.Equal(t, "ssh -i id_rsa -p 2222 -l root localhost", utils.StripStyles(model.Title))
}

func Test_handleKeyboardEvent_cancelWhileFiltering(t *testing.T) {
	// Test that when user presses 'Esc' key while filtering, the model doesn't lose focus
	// Create model
	model := NewMockListModel(false)
	model.Init()

	// Make sure there are 3 items in the collection
	require.Len(t, model.VisibleItems(), 3)

	// Check that first item is selected
	require.IsType(t, ListItemHost{}, model.SelectedItem())
	require.Equal(t, "Mock Host 1", model.SelectedItem().(ListItemHost).Title())

	// Check that current status is "Unfiltered" and then enter filtering mode
	require.Equal(t, list.Unfiltered, model.FilterState())
	model.Update(tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune{'/'},
	})
	require.Equal(t, list.Filtering, model.FilterState())

	// When in filter mode type '2', so only "Mock Host 2" will become visible
	_, cmds := model.Update(tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune{'2'},
	})

	// Extract batch messages returned by the model
	msgs := []tea.Msg{}
	testutils.CmdToMessage(cmds, &msgs)

	// Send those messages back to the model
	for _, m := range msgs {
		model.Update(m)
	}

	require.Len(t, model.VisibleItems(), 1)

	// When in filter mode type 'Esc', and ensure that we exited filter mode but the
	// focus is set on the first item from the search results
	_, cmds = model.Update(tea.KeyMsg{Type: tea.KeyEscape})
	// Extract batch messages returned by the model
	testutils.CmdToMessage(cmds, &msgs)

	// Send those messages back to the model
	for _, m := range msgs {
		model.Update(m)
	}

	require.Equal(t, list.Unfiltered, model.FilterState())
	require.Len(t, model.VisibleItems(), 3)
	// By triggering filter, though we haven't selected anything, we implicitly selected the first item from the search results
	require.Equal(t, "Mock Host 2", model.SelectedItem().(ListItemHost).Title())
}

func Test_handleKeyboardEvent_clearFilter(t *testing.T) {
	// When enable filter and then focus a host (while still in filtering mode)
	// and after that press Escape button (to exit from filter mode),
	// we must ensure that focus hasn't changed and the same host is still
	// focused which was selected when filter was enabled.
	model := NewMockListModel(false)
	model.Init()

	// Make sure there are 3 items in the collection
	require.Len(t, model.VisibleItems(), 3)

	// Check that first item is selected
	require.IsType(t, ListItemHost{}, model.SelectedItem())
	require.Equal(t, "Mock Host 1", model.SelectedItem().(ListItemHost).Title())

	// Check that current status is "Unfiltered" and then enter filtering mode
	require.Equal(t, list.Unfiltered, model.FilterState())
	model.Update(tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune{'/'},
	})
	require.Equal(t, list.Filtering, model.FilterState())

	// When in filter mode type '2', so only "Mock Host 2" will become visible
	_, cmds := model.Update(tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune{'2'},
	})

	// Extract batch messages returned by the model
	msgs := []tea.Msg{}
	testutils.CmdToMessage(cmds, &msgs)

	// Send those messages back to the model
	for _, m := range msgs {
		model.Update(m)
	}

	require.Len(t, model.VisibleItems(), 1)

	// When in filter mode press 'Enter', and ensure that
	// focus is set to host "Mock Host 2"
	_, cmds = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	// Extract batch messages returned by the model
	testutils.CmdToMessage(cmds, &msgs)
	for _, m := range msgs {
		model.Update(m)
	}
	// Check that we're still in filter mode when pressed Enter
	require.Equal(t, list.FilterApplied, model.FilterState())
	require.Len(t, model.VisibleItems(), 1)
	// And check that now "Mock Host 2" is selected
	require.Equal(t, "Mock Host 2", model.SelectedItem().(ListItemHost).Title())

	// Now press 'Esc', and ensure that we exited filter mode but the
	// focus is set on the same item which was focused while we were in filter mode
	_, cmds = model.Update(tea.KeyMsg{Type: tea.KeyEscape})
	// Extract batch messages returned by the model
	testutils.CmdToMessage(cmds, &msgs)
	for _, m := range msgs {
		model.Update(m)
	}
	require.Len(t, model.VisibleItems(), 3)
	require.Equal(t, list.Unfiltered, model.FilterState())
	require.Equal(t, "Mock Host 2", model.SelectedItem().(ListItemHost).Title())
}

func Test_handleKeyboardEvent_selectGroup(t *testing.T) {
	model := NewMockListModel(false)
	model.Init()
	_, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'z'}})
	res := cmd()
	require.IsType(t, message.OpenViewSelectGroup{}, res)
}

func Test_handleKeyboardEvent_connect(t *testing.T) {
	// Check that when we press Enter button while host is selected
	// we dispatch processConstruct command from the host list model
	model := NewMockListModel(false)
	model.Init()

	// Make sure there are 3 items in the collection
	require.Len(t, model.VisibleItems(), 3)
	// Check that first item is selected
	require.IsType(t, ListItemHost{}, model.SelectedItem())
	require.Equal(t, "Mock Host 1", model.SelectedItem().(ListItemHost).Title())

	// Hit enter
	_, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	require.IsType(t, message.RunProcessSSHConnect{}, cmd())
}

func Test_handleKeyboardEvent_copyID(t *testing.T) {
	// Just check that we enter copyID mode when a host is selected and press "t" button
	model := NewMockListModel(false)
	model.Init()

	// Make sure there are 3 items in the collection
	require.Len(t, model.VisibleItems(), 3)
	// Check that first item is selected
	require.IsType(t, ListItemHost{}, model.SelectedItem())
	require.Equal(t, "Mock Host 1", model.SelectedItem().(ListItemHost).Title())

	_, cmds := model.Update(tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune{'t'},
	})

	msgs := []tea.Msg{}
	testutils.CmdToMessage(cmds, &msgs)
	for _, m := range msgs {
		model.Update(m)
	}

	require.Equal(t, modeSSHCopyID, model.mode)
}

func Test_handleKeyboardEvent_remove(t *testing.T) {
	// Just check that we enter removeItem mode when a host is selected and press "t" button
	model := NewMockListModel(false)
	model.Init()

	// Make sure there are 3 items in the collection
	require.Len(t, model.VisibleItems(), 3)
	// Check that first item is selected
	require.IsType(t, ListItemHost{}, model.SelectedItem())
	require.Equal(t, "Mock Host 1", model.SelectedItem().(ListItemHost).Title())

	_, cmds := model.Update(tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune{'x'},
	})

	msgs := []tea.Msg{}
	testutils.CmdToMessage(cmds, &msgs)
	for _, m := range msgs {
		model.Update(m)
	}

	require.Equal(t, modeRemoveItem, model.mode)
}

func Test_handleKeyboardEvent_edit(t *testing.T) {
	t.Skip("In progress")
}

func Test_handleKeyboardEvent_append(t *testing.T) {
	t.Skip("In progress")
}

func Test_constructProcessCmd(t *testing.T) {
	// Test that we receive expected messages when invoke constructProcessCmd function
	lm := *NewMockListModel(false)
	lm.Init()
	connectSSHResultCmd := lm.constructProcessCmd(constant.ProcessTypeSSHConnect)
	selectedHost := lm.SelectedItem().(ListItemHost).Host
	require.Equal(t, message.RunProcessSSHConnect{Host: selectedHost}, connectSSHResultCmd())

	sshCopyID := lm.constructProcessCmd(constant.ProcessTypeSSHCopyID)
	require.Equal(t, message.RunProcessSSHCopyID{Host: selectedHost}, sshCopyID())
}

func TestUpdate_SearchFunctionOfInnerModelIsNotRegressed(t *testing.T) {
	// Test that filtering is working properly

	// Create mock storage which contains hosts:
	// "Mock Host 1"
	// "Mock Host 2"
	// "Mock Host 3"
	storage := testutils.NewMockStorage(false)
	fakeAppState := state.Application{Selected: 1}

	// Create model
	model := New(context.TODO(), storage, &fakeAppState, &mocklogger.Logger{})
	model.logger = &mocklogger.Logger{}
	model.Init()

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
	testutils.CmdToMessage(cmds, &msgs)

	// Feed all messages one by one to the model
	for _, msg := range msgs {
		model.Update(msg)
	}

	// Ensure, that only one item left in the list (which is "Mock Host 1")
	require.Len(t, model.VisibleItems(), 1)
}

func TestUpdate_ToggleBetweenScreenLayouts(t *testing.T) {
	// Create mock storage which contains hosts:
	// "Mock Host 1"
	// "Mock Host 2"
	// "Mock Host 3"
	storage := testutils.NewMockStorage(false)
	fakeAppState := state.Application{Selected: 1}

	// Create model
	model := New(context.TODO(), storage, &fakeAppState, &mocklogger.Logger{})
	model.logger = &mocklogger.Logger{}
	model.Init()

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

	fakeAppState.ScreenLayout = constant.ScreenLayoutCompact
	// Ensure that screen layout is equal to
	require.Equal(t, constant.ScreenLayoutCompact, fakeAppState.ScreenLayout)

	// Toggle layout again and check that it's now set to "normal"
	model.Update(tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune{'v'},
	})
	require.Equal(t, constant.ScreenLayoutDescription, fakeAppState.ScreenLayout)

	// Toggle layout again and check that it's now set to "group"
	model.Update(tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune{'v'},
	})
	require.Equal(t, constant.ScreenLayoutGroup, fakeAppState.ScreenLayout)
}

func Test_HandleKeyboardEvent_Escape(t *testing.T) {
	// If group is selected and type Escape key, the model
	// should dispatch open group view message
	model := NewMockListModel(false)
	model.appState.Group = "Group 1"
	_, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	require.IsType(t, message.OpenViewSelectGroup{}, cmd())

	// If group is NOT selected and press Escape key, the app
	// should ask the user whether it wants to close the program
	model = NewMockListModel(false)
	model.appState.Group = ""
	_, cmd = model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	require.Nil(t, cmd)
	require.Equal(t, modeCloseApp, model.mode)
	require.Equal(t, "close app? (y/N)", utils.StripStyles(model.Title))
}

func TestUpdate_HostFocusPreservedAfterClearFilterMessage(t *testing.T) {
	// Test that the same host is selected after we exit filter mode (clear filter with "Esc" button)

	// Create mock storage which contains hosts:
	// "Mock Host 1"
	// "Mock Host 2"
	// "Mock Host 3"
	storage := testutils.NewMockStorage(false)
	fakeAppState := state.Application{Selected: 1}

	// Create model
	model := New(context.TODO(), storage, &fakeAppState, &mocklogger.Logger{})
	model.logger = &mocklogger.Logger{}
	model.Init()

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
	testutils.CmdToMessage(cmds, &msgs)

	// Send those messages back to the model
	for _, m := range msgs {
		model.Update(m)
	}

	assert.Len(t, model.VisibleItems(), 1)
}

// ==============================================
// ============== utility methods ===============
// ==============================================

func NewMockListModel(storageShouldFail bool) *listModel {
	storage := testutils.NewMockStorage(storageShouldFail)
	mockState := state.Application{Selected: 1}

	// Create listModel using constructor function (using 'New' is important to preserve hotkeys)
	lm := New(context.TODO(), storage, &mockState, &mocklogger.Logger{})

	items := make([]list.Item, 0)
	// Wrap hosts into List items
	hosts := storage.Hosts
	for _, h := range hosts {
		items = append(items, ListItemHost{Host: h})
	}

	lm.SetItems(items)

	return lm
}
