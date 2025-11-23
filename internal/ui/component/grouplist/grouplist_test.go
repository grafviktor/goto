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

func TestHandleKeyboardEvent_Enter(t *testing.T) {
	listModel := NewMockGroupModel(false)
	listModel.loadItems()

	// Test case 1: Can handle "no group" selection
	require.Equal(t, noGroupSelected, listModel.SelectedItem().(ListItemHostGroup).Title())
	_, cmd := listModel.Update(tea.KeyMsg{Type: tea.KeyEnter})
	// Press enter on "no group" item
	var actualMsgs1 []tea.Msg
	testutils.CmdToMessage(cmd, &actualMsgs1)
	expectedMsgs := []tea.Msg{
		message.GroupSelect{Name: ""},
		message.ViewGroupListClose{},
	}
	require.ElementsMatch(t, expectedMsgs, actualMsgs1)
	require.Equal(t, noGroupSelected, listModel.SelectedItem().(ListItemHostGroup).Title())

	// Test case 2: Can handle "Group 1" selection (which is a manually created group in mock storage)
	listModel.Update(tea.KeyMsg{Type: tea.KeyDown})
	_, cmd = listModel.Update(tea.KeyMsg{Type: tea.KeyEnter})
	var actualMsgs2 []tea.Msg
	testutils.CmdToMessage(cmd, &actualMsgs2)
	expectedMsgs = []tea.Msg{
		message.GroupSelect{Name: "Group 1"},
		message.ViewGroupListClose{},
	}

	require.ElementsMatch(t, expectedMsgs, actualMsgs2)
	require.Equal(t, "Group 1", listModel.SelectedItem().(ListItemHostGroup).Title())

	// Test case 3: The model is in filter mode
	t.Fail()
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
