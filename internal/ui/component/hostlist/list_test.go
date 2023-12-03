// Package hostlist implements the host list view.
package hostlist

import (
	"testing"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/grafviktor/goto/internal/model"
	"github.com/stretchr/testify/require"
)

func Test_ListTitleUpdate(t *testing.T) {
	t.Run("Focus Changed Message", func(t *testing.T) {
		// Create a new host
		h := model.NewHost(0, "", "", "localhost", "root", "id_rsa", "2222")

		// Create items
		items := []list.Item{ListItemHost{h}}

		// Create a lm with initial state
		lm := listModel{innerModel: list.New(items, list.NewDefaultDelegate(), 0, 0)}

		// Select host
		lm.innerModel.Select(0)

		// Create a message of type msgFocusChanged
		msg := msgFocusChanged{}
		// Apply the function
		lm = lm.listTitleUpdate(msg)

		require.Equal(t, lm.innerModel.Title, "ssh localhost -l root -p 2222 -i id_rsa")
	})
}

func Test_listModel_Change_Selection(t *testing.T) {
	getListModel := func() *listModel {
		// Create a new host
		h := model.NewHost(0, "", "", "localhost", "root", "id_rsa", "2222")

		// Add three items to the list
		items := []list.Item{ListItemHost{h}, ListItemHost{h}, ListItemHost{h}, ListItemHost{h}}

		// Create listModel using constructor function (using 'New' is important to preserve hotkeys)
		lm := New(nil, nil, nil, nil)
		lm.innerModel.SetItems(items)

		// Select item at index 1. We need this preselection in order to test 'focus previous' and 'focus next' messages
		lm.innerModel.Select(1)

		return &lm
	}

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
			*getListModel(),
			tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}},
		},
		{
			"Select next using '↓' key",
			2,
			*getListModel(),
			tea.KeyMsg{Type: tea.KeyDown},
		},
		{
			"Select next using 'tab' key",
			2,
			*getListModel(),
			tea.KeyMsg{Type: tea.KeyTab},
		},
		// Simulate focus previous event
		{
			"Select previous using 'k' key",
			0,
			*getListModel(),
			tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}},
		},
		{
			"Select previous using '↑' key",
			0,
			*getListModel(),
			tea.KeyMsg{Type: tea.KeyUp},
		},
		{
			"Select previous using 'shift+tab' key",
			0,
			*getListModel(),
			tea.KeyMsg{Type: tea.KeyShiftTab},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
