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
	require.False(t, model.FilteringEnabled())
	require.False(t, model.ShowStatusBar())
	require.False(t, model.FilteringEnabled())
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
	listModel.Update(message.OpenViewSelectGroup{})

	// Selected group is "no group"
	require.Equal(t, noGroupSelected, listModel.SelectedItem().(ListItemHostGroup).Title())
}

func TestHandleKeyboardEvent_Enter(t *testing.T) {
	// Can Select group
	listModel := NewMockGroupModel(false)
	listModel.loadItems()
	require.Equal(t, noGroupSelected, listModel.SelectedItem().(ListItemHostGroup).Title())

	// Select Group 1
	listModel.Update(tea.KeyMsg{Type: tea.KeyDown})
	_, cmd := listModel.Update(tea.KeyMsg{Type: tea.KeyEnter})

	var actualMsgs []tea.Msg
	testutils.CmdToMessage(cmd, &actualMsgs)
	expectedMsgs := []tea.Msg{
		message.GroupSelected{Name: "Group 1"},
		message.CloseViewSelectGroup{},
	}

	require.ElementsMatch(t, expectedMsgs, actualMsgs)
	require.Equal(t, "Group 1", listModel.SelectedItem().(ListItemHostGroup).Title())
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
		message.GroupSelected{Name: ""},
		message.CloseViewSelectGroup{},
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

func NewMockGroupModel(storageShouldFail bool) *model {
	mockState := state.Application{Selected: 1}
	storage := testutils.NewMockStorage(storageShouldFail)
	return New(context.TODO(), storage, &mockState, &mocklogger.Logger{})
}
