package grouplist

import (
	"context"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/require"

	"github.com/grafviktor/goto/internal/state"
	testutils "github.com/grafviktor/goto/internal/testutils"
	"github.com/grafviktor/goto/internal/testutils/mocklogger"
	"github.com/grafviktor/goto/internal/ui/message"
)

func TestNew(t *testing.T) {
	model := NewMockGroupModel(false)
	require.True(t, model.FilteringEnabled())
	// Quit app keys is disabled
	require.False(t, model.KeyMap.Quit.Enabled())
	require.False(t, model.KeyMap.ForceQuit.Enabled())
	require.Equal(t, "select group", model.Title)
}

func TestInit(t *testing.T) {
	model := NewMockGroupModel(false)
	cmd := model.Init()
	require.IsType(t, tea.Cmd(nil), cmd)
}

func TestUpdate(t *testing.T) {
	listModel := NewMockGroupModel(false)

	// Can handle tea.WindowSizeMsg
	require.Equal(t, 0, listModel.Height())
	require.Equal(t, 0, listModel.Width())
	windowSizeMsg := tea.WindowSizeMsg{Width: 100, Height: 100}
	listModel.Update(windowSizeMsg)
	require.Positive(t, listModel.Height())
	require.Positive(t, listModel.Width())

	// Loads hosts when the form is shown
	listModel = NewMockGroupModel(false)
	listModel.Update(message.ViewGroupListOpen{})

	// Selected group is "no group"
	require.Equal(t, noGroupSelected, listModel.SelectedItem().(ListItemHostGroup).Title())
}

func Test_handleKeyboardEvent(t *testing.T) {
	listModel := NewMockGroupModel(false)
	listModel.loadItems()

	tests := []struct {
		name      string
		keyMsg    tea.KeyMsg
		expectCmd bool
	}{
		{
			name:      "Can handle Enter key",
			keyMsg:    tea.KeyMsg{Type: tea.KeyEnter},
			expectCmd: true,
		},
		{
			name:      "Can handle Esc key",
			keyMsg:    tea.KeyMsg{Type: tea.KeyEsc},
			expectCmd: true,
		},
		{
			name:      "Unhandled key returns nil cmd",
			keyMsg:    tea.KeyMsg{Type: tea.KeyDown},
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
}

func Test_handleEnterKey(t *testing.T) {
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

	listModel := NewMockGroupModel(false)
	listModel.loadItems()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			listModel.Select(tt.itemIndex)
			cmd := listModel.handleEnterKey()

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
	// When pressing Enter key while filtering, it should return nil cmd
	listModel := NewMockGroupModel(false)
	listModel.loadItems()
	listModel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")})

	require.True(t, listModel.SettingFilter()) // Activate filter mode
	_, cmd := listModel.Update(tea.KeyMsg{Type: tea.KeyEnter})
	require.Nil(t, cmd)
}

func TestHandleKeyboardEvent_Esc(t *testing.T) {
	// Can handle Esc key
	listModel := NewMockGroupModel(false)
	listModel.loadItems()
	require.Equal(t, noGroupSelected, listModel.SelectedItem().(ListItemHostGroup).Title())

	// Select Group 1
	listModel.Update(tea.KeyMsg{Type: tea.KeyDown})
	listModel.Update(tea.KeyMsg{Type: tea.KeyEnter})
	require.Equal(t, "Group 1", listModel.SelectedItem().(ListItemHostGroup).Title())

	// Now press Escape key and ensure that Model will send group unselect message and closed the form
	_, cmd := listModel.Update(tea.KeyMsg{Type: tea.KeyEsc})
	var actualMsgs []tea.Msg
	testutils.CmdToMessage(cmd, &actualMsgs)

	expectedMsgs := []tea.Msg{
		message.GroupSelect{Name: ""},
		message.ViewGroupListClose{},
	}

	// Note that though, Escape key was pressed, the group remains selected inside the group list component.
	// This is by design - if user opens group list dialog again, previously selected group will be focused.
	require.Equal(t, "Group 1", listModel.SelectedItem().(ListItemHostGroup).Title())
	require.ElementsMatch(t, expectedMsgs, actualMsgs)
}

func TestLoadItems(t *testing.T) {
	// Storage works fine
	model := NewMockGroupModel(false)
	model.loadItems()
	require.Equal(t, noGroupSelected, model.Items()[0].(ListItemHostGroup).Title())
	require.Len(t, model.Items(), 4) // 1 "No Group" item + 3 items from database

	// Storage fails
	model = NewMockGroupModel(true)
	cmd := model.loadItems()
	require.IsType(t, message.ErrorOccurred{}, cmd())
	require.Empty(t, model.Items())
}

// ==============================================
// ============== utility methods ===============
// ==============================================

func NewMockGroupModel(storageShouldFail bool) *Model {
	mockState := state.State{Selected: 1}
	storage := testutils.NewMockStorage(storageShouldFail)
	return New(context.TODO(), storage, &mockState, &mocklogger.Logger{})
}
