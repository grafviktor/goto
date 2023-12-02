// Package hostlist implements the host list view.
package hostlist

import (
	"errors"
	"testing"

	"github.com/charmbracelet/bubbles/list"
	"github.com/grafviktor/goto/internal/model"
	"github.com/stretchr/testify/require"
)

func TestListTitleUpdate(t *testing.T) {
	// List title should show error message if error happened
	t.Run("Error Message", func(t *testing.T) {
		errMsg := errors.New("test error")
		msg := msgErrorOccured{err: errMsg}
		model := listModel{innerModel: list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)}
		newModel := model.listTitleUpdate(msg)

		if newModel.innerModel.Title != errMsg.Error() {
			t.Errorf("Expected title to be %q, but got %q", errMsg.Error(), newModel.innerModel.Title)
		}
	})

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
