package grouplist

import (
	"context"
	"reflect"
	"testing"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/grafviktor/goto/internal/state"
	testutils "github.com/grafviktor/goto/internal/testutils"
	"github.com/grafviktor/goto/internal/testutils/mocklogger"
	"github.com/grafviktor/goto/internal/ui/message"
)

func TestNew(t *testing.T) {
	model := newMockModel(false)
	require.True(t, model.FilteringEnabled())
	// Quit app keys is disabled
	require.False(t, model.KeyMap.Quit.Enabled())
	require.False(t, model.KeyMap.ForceQuit.Enabled())
	require.Equal(t, "select group", model.Title)
}

func TestInit(t *testing.T) {
	model := newMockModel(false)
	var msgs []tea.Msg
	testutils.CmdToMessage(model.Init(), &msgs)

	require.Equal(t, message.GroupListLoadComplete{GroupCount: 3}, msgs[len(msgs)-1])
	require.Len(t, model.Items(), 4)
}

func TestUpdate(t *testing.T) {
	listModel := newMockModel(false)

	// Can handle tea.WindowSizeMsg
	require.Equal(t, 0, listModel.Height())
	require.Equal(t, 0, listModel.Width())
	windowSizeMsg := tea.WindowSizeMsg{Width: 100, Height: 100}
	listModel.Update(windowSizeMsg)
	require.Positive(t, listModel.Height())
	require.Positive(t, listModel.Width())

	// Loads hosts when the form is shown
	listModel = newMockModel(false)
	listModel.Update(message.ViewGroupListOpen{})

	// Selected group is "no group"
	require.Equal(t, noGroupSelected, listModel.SelectedItem().(ListItemHostGroup).Title())
}

func Test_handleKeyboardEvent(t *testing.T) {
	t.Run("General keyboard event test", func(t *testing.T) {
		listModel := newMockModel(false)
		listModel.loadItems()

		tests := []struct {
			name      string
			keyMsg    tea.KeyPressMsg
			expectCmd bool
		}{
			{
				name:      "Can handle Enter key",
				keyMsg:    tea.KeyPressMsg{Code: tea.KeyEnter},
				expectCmd: true,
			},
			{
				name:      "Can handle Esc key",
				keyMsg:    tea.KeyPressMsg{Code: tea.KeyEsc},
				expectCmd: true,
			},
			{
				name:      "Unhandled key returns nil cmd",
				keyMsg:    tea.KeyPressMsg{Code: tea.KeyDown},
				expectCmd: false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				cmd := listModel.handleKeyboardEvent(tt.keyMsg)
				if tt.expectCmd {
					require.NotNil(t, cmd)
				} else {
					require.Nil(t, cmd)
				}
			})
		}
	})

	t.Run("Quit app keyboard event test", func(t *testing.T) {
		tests := map[string]struct {
			keyMsg          tea.KeyPressMsg
			isFilterEnabled bool
			expectCmd       tea.Cmd
		}{
			"Type 'q'": {
				keyMsg:          tea.KeyPressMsg{Code: 'q'},
				isFilterEnabled: false,
				expectCmd:       tea.Quit,
			},
			"Type 'q' when filtering groups": {
				keyMsg:          tea.KeyPressMsg{Code: 'q'},
				isFilterEnabled: true,
				expectCmd:       nil,
			},
		}

		for testName, tt := range tests {
			t.Run(testName, func(t *testing.T) {
				// We mutate model state, so let's re-create it in every test.
				listModel := newMockModel(false)
				listModel.loadItems()

				if tt.isFilterEnabled {
					listModel.SetFilterState(list.Filtering)
				}
				cmd := listModel.handleKeyboardEvent(tt.keyMsg)
				if tt.expectCmd == nil {
					require.Nil(t, cmd)
				} else {
					require.Equal(t, tt.expectCmd(), cmd())
				}
			})
		}
	})
}

func Test_handleSelectGroup(t *testing.T) {
	tests := []struct {
		name         string
		group        string
		itemIndex    int
		expectedMsgs []tea.Msg
	}{
		{
			name:      "Can handle 'no group' selection",
			group:     noGroupSelected,
			itemIndex: 0,
			expectedMsgs: []tea.Msg{
				message.GroupSelect{Name: ""},
				message.ViewGroupListClose{},
			},
		},
		{
			name:      "Can handle Group 1 selection ",
			group:     "Group 1",
			itemIndex: 1,
			expectedMsgs: []tea.Msg{
				message.GroupSelect{Name: "Group 1"},
				message.ViewGroupListClose{},
			},
		},
		{
			name:         "Can handle empty group list ",
			group:        "",
			itemIndex:    55, // Out of range index, selected item will be nil
			expectedMsgs: nil,
		},
	}

	listModel := newMockModel(false)
	listModel.loadItems()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			listModel.Select(tt.itemIndex)
			cmd := listModel.handleSelectGroup()

			var actualMsgs []tea.Msg
			testutils.CmdToMessage(cmd, &actualMsgs)
			require.ElementsMatch(t, tt.expectedMsgs, actualMsgs)
			if listModel.SelectedItem() != nil {
				require.Equal(t, tt.group, listModel.SelectedItem().(ListItemHostGroup).Title())
			}
		})
	}
}

func Test_handleEnterKey_WhenFiltering(t *testing.T) {
	// When pressing Enter key while filtering, it should return nil cmd, as we do not want to select
	// the first item in the list, but instead let user to select an item using Up/Down keys.
	listModel := newMockModel(false)
	listModel.loadItems()
	// Put the list in filter mode
	listModel.Update(tea.KeyPressMsg{Code: '/'})

	require.True(t, listModel.SettingFilter()) // Activate filter mode
	_, cmd := listModel.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	require.Nil(t, cmd)
}

func Test_handleEscapeKey(t *testing.T) {
	// Test case 1: Press escape in filter mode
	// Can handle Esc key
	listModel := newMockModel(false)
	listModel.loadItems()
	// Put the list in filter mode
	listModel.Update(tea.KeyPressMsg{Code: '/'})
	require.True(t, listModel.SettingFilter()) // Verify that filter mode is activate
	_, cmd := listModel.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	require.Nil(t, cmd)

	// Test case 2: Press escape when not in filter mode - it must deselect the group and close the form
	require.False(t, listModel.SettingFilter()) // Verify that we're not in filter mode after the first test case
	_, cmd = listModel.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	var actualMsgs []tea.Msg
	testutils.CmdToMessage(cmd, &actualMsgs)
	require.ElementsMatch(t, actualMsgs, []tea.Msg{
		message.GroupSelect{Name: ""},
		message.ViewGroupListClose{},
	})
}

func TestLoadItems(t *testing.T) {
	// Storage works fine
	model := newMockModel(false)
	model.loadItems()
	require.Equal(t, noGroupSelected, model.Items()[0].(ListItemHostGroup).Title())
	require.Len(t, model.Items(), 4) // 1 "No Group" item + 3 items from database
	require.Equal(t, noGroupSelected, model.SelectedItem().(ListItemHostGroup).Title())

	// Previously selected group is focused after reload.
	model.appState.Group = "group 2"
	model.loadItems()
	require.Equal(t, "Group 2", model.SelectedItem().(ListItemHostGroup).Title())

	// A group that no longer exists falls back to "~ all ~".
	model.appState.Group = "missing"
	model.loadItems()
	require.Equal(t, noGroupSelected, model.SelectedItem().(ListItemHostGroup).Title())

	// Storage fails
	model = newMockModel(true)
	cmd := model.loadItems()
	require.IsType(t, message.ErrorOccurred{}, cmd())
	require.Empty(t, model.Items())
}

func Test_handleMouseMsg(t *testing.T) {
	mouseMsg := tea.MouseClickMsg{
		Button: tea.MouseLeft,
		// A hack - we must simulate a click below hostlist header. I just know that header's height is 4.
		Y: 5,
	}

	var model *Model
	beforeEach := func() {
		model = newMockModel(false)
		model.loadItems()
	}

	t.Run("Grouplist is in filter mode, mouse message should be ignored", func(t *testing.T) {
		beforeEach()
		model.SetFilterState(list.Filtering)
		cmd := model.handleMouseMsg(mouseMsg)
		require.Nil(t, cmd)
	})

	t.Run("Grouplist receives right mouse click, the message should be ignored", func(t *testing.T) {
		beforeEach()
		rightMouseMsg := mouseMsg
		rightMouseMsg.Button = tea.MouseRight
		cmd := model.handleMouseMsg(rightMouseMsg)
		require.Nil(t, cmd)
	})

	t.Run("Finally grouplist receives left mouse click and returns some tea.Cmd", func(t *testing.T) {
		beforeEach()
		cmd := model.handleMouseMsg(mouseMsg)
		require.NotNil(t, cmd) // Failing because there is no visible items in the list. Requires debugging.
	})
}

func Test_focusItem(t *testing.T) {
	var model *Model
	beforeEach := func() {
		model = newMockModel(false)
		model.SetHeight(10) // Required to display more than one item on the first page
		model.loadItems()
	}

	t.Run("Click on paginator", func(t *testing.T) {
		beforeEach()
		model.SetHeight(5) // Decrease list height to simulate more than one page.
		paginatorPosition := model.paginatorRowPosition()
		require.True(t, model.Paginator.OnFirstPage())
		model.focusItem(paginatorPosition)
		// We no longer on the first page after clicking on paginator.
		require.False(t, model.Paginator.OnFirstPage())
	})

	t.Run("Click above the first item must be ignored", func(t *testing.T) {
		beforeEach()
		cmd := model.focusItem(4) // Click above the first item, assuming header height is 4.
		require.Nil(t, cmd)
	})

	t.Run("First click on an item must focus it and start timer for second click", func(t *testing.T) {
		beforeEach()
		assert.Equal(t, 0, model.Index())

		// Assuming header height is 4. Click on the second item (index=1).
		cmd := model.focusItem(6)
		require.True(t, model.waitForSecondClick)
		require.Equal(t, 1, model.Index()) // The second item should be selected.
		require.Equal(t, "msgSecondClickTimeOut", reflect.TypeOf(cmd()).Name())
	})

	t.Run("Second click on an item must open it and close group list", func(t *testing.T) {
		beforeEach()
		model.waitForSecondClick = true

		cmd := model.focusItem(5)
		var msgs []tea.Msg
		testutils.CmdToMessage(cmd, &msgs)
		var msgNames []string
		for _, msg := range msgs {
			msgNames = append(msgNames, reflect.TypeOf(msg).Name())
		}
		require.Contains(t, msgNames, "GroupSelect")
		require.Contains(t, msgNames, "ViewGroupListClose")
	})
}

func Test_handlePaginatorClick(t *testing.T) {
	var model *Model
	beforeEach := func() {
		model = newMockModel(false)
		model.SetHeight(15) // 15 rows is enough to display 4 items (3 from DB +  "~ all ~") on a single page.
		model.loadItems()
		model.waitForSecondClick = true
	}

	t.Run("If there is only a single page, then paginator click must be ignored", func(t *testing.T) {
		beforeEach()
		require.Equal(t, 1, model.Paginator.TotalPages)
		model.handlePaginatorClick()
		// Model must not reset click flag, when paginator click is ignored.
		require.True(t, model.waitForSecondClick)
		// Because there is only a single page, we're on first and last page at the same time.
		require.True(t, model.Paginator.OnFirstPage())
	})

	t.Run("Must go to next page if there is more than one page", func(t *testing.T) {
		beforeEach()
		model.SetHeight(8) // Decrease model height, it will give us 2 pages
		require.Equal(t, 2, model.Paginator.TotalPages)
		model.handlePaginatorClick()
		// Model must reset click flag, when move to the next page.
		require.False(t, model.waitForSecondClick)
		// Because there is only a single page, we're on first and last page at the same time.
		require.False(t, model.Paginator.OnFirstPage())
	})

	t.Run("Must go to the first page when on last page and someone clicks on paginator", func(t *testing.T) {
		beforeEach()
		model.SetHeight(8) // Decrease model height, it will give us 2 pages
		require.Equal(t, 2, model.Paginator.TotalPages)
		model.handlePaginatorClick()
		require.True(t, model.Paginator.OnLastPage()) // Now we're on the last page.

		model.handlePaginatorClick() // Click one again
		require.False(t, model.waitForSecondClick)
		require.True(t, model.Paginator.OnFirstPage()) // And ensure that we're on the first page.
	})
}

func Test_itemAt(t *testing.T) {
	model := newMockModel(false)
	model.SetHeight(15) // Required to display more than one item on the first page
	model.loadItems()

	// Header offset is 4. The list has FOUR items: "~ all ~" plus 3 from the database,
	// so a click at Y=9 is the first row below the last item.
	tests := []struct {
		name           string
		mousePositionY int
		want           int
	}{
		{
			name:           "Click above the first item must return negative index",
			mousePositionY: 4,
			want:           -1,
		},
		{
			name:           "Click on the first item must return index 0",
			mousePositionY: 5,
			want:           0,
		},
		{
			name:           "Click on the second item must return index 1",
			mousePositionY: 6,
			want:           1,
		},
		{
			name:           "Click below the last item must return negative index",
			mousePositionY: 9,
			want:           -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, model.itemAt(tt.mousePositionY))
		})
	}
}

// ==============================================
// ============== utility methods ===============
// ==============================================

func newMockModel(storageShouldFail bool) *Model {
	mockState := state.State{Selected: 1}
	storage := testutils.NewMockStorage(storageShouldFail)
	return New(context.TODO(), storage, &mockState, &mocklogger.Logger{})
}
