// Package edithost contains UI components for editing host model attributes.
package edithost

import (
	"context"
	"fmt"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/grafviktor/goto/internal/mock"
	"github.com/grafviktor/goto/internal/state"
)

func Test_NotEmptyValidator(t *testing.T) {
	tests := []struct {
		input    string
		expected error
	}{
		{"", fmt.Errorf("value is required")},
		{"non-empty", nil},
	}

	for _, test := range tests {
		result := notEmptyValidator(test.input)

		if (result == nil && test.expected != nil) || (result != nil && test.expected == nil) {
			t.Errorf("For input %q, expected error %v but got %v", test.input, test.expected, result)
		}

		if result != nil && result.Error() != test.expected.Error() {
			t.Errorf("For input %q, expected error %v but got %v", test.input, test.expected, result)
		}
	}
}

func Test_NetworkPortValidator(t *testing.T) {
	tests := []struct {
		input    string
		expected error
	}{
		{"", nil},
		{"abc", fmt.Errorf("network port must be a number which is less than 65,535")},
		{"0", fmt.Errorf("network port must be a number which is less than 65,535")},
		{"65536", fmt.Errorf("network port must be a number which is less than 65,535")},
		{"123", nil},
	}

	for _, test := range tests {
		result := networkPortValidator(test.input)

		if (result == nil && test.expected != nil) || (result != nil && test.expected == nil) {
			t.Errorf("For input %q, expected error %v but got %v", test.input, test.expected, result)
		}

		if result != nil && result.Error() != test.expected.Error() {
			t.Errorf("For input %q, expected error %v but got %v", test.input, test.expected, result)
		}
	}
}

func Test_GetKeyMap(t *testing.T) {
	// When title or address is selected, we can copy its values between each other using a shortcut
	keyMap := getKeyMap(inputTitle)
	require.True(t, keyMap.CopyInputValue.Enabled())
	keyMap = getKeyMap(inputTitle)
	require.True(t, keyMap.CopyInputValue.Enabled())

	// However, when any other input selected, this keyboard shortcut should NOT be available.
	keyMap = getKeyMap(inputDescription)
	require.False(t, keyMap.CopyInputValue.Enabled())
}

// func (m editModel) save(_ tea.Msg) (editModel, tea.Cmd) {
// 	for i := range m.inputs {
// 		if m.inputs[i].Validate != nil {
// 			if err := m.inputs[i].Validate(m.inputs[i].Value()); err != nil {
// 				m.inputs[i].Err = err
// 				m.title = fmt.Sprintf("%s is not valid", m.inputs[i].Label)

// 				return m, nil
// 			}
// 		}

// 		switch i {
// 		case inputTitle:
// 			m.host.Title = m.inputs[i].Value()
// 		case inputAddress:
// 			m.host.Address = m.inputs[i].Value()
// 		case inputDescription:
// 			m.host.Description = m.inputs[i].Value()
// 		case inputLogin:
// 			m.host.LoginName = m.inputs[i].Value()
// 		case inputNetworkPort:
// 			m.host.RemotePort = m.inputs[i].Value()
// 		case inputIdentityFile:
// 			m.host.PrivateKeyPath = m.inputs[i].Value()
// 		}
// 	}

// 	host, _ := m.hostStorage.Save(m.host)
// 	return m, tea.Batch(
// 		message.TeaCmd(MsgClose{}),
// 		// Order matters here! 'HostListSelectItem' message should be dispatched
// 		// before 'MsgRepoUpdated'. The reasons of that is because
// 		// 'MsgRepoUpdated' handler automatically sets focus on previously selected item.
// 		message.TeaCmd(message.HostListSelectItem{HostID: host.ID}),
// 		message.TeaCmd(hostlist.MsgRepoUpdated{}),
// 	)
// }

func Test_save(t *testing.T) {
	state := state.ApplicationState{}

	editHostModel := New(context.TODO(), mock.NewMockStorage(true), &state, &mock.MockLogger{})
	require.Equal(t, inputTitle, editHostModel.focusedInput)

	editHostModel.inputs[inputDescription].SetValue("test")
	editHostModel.inputs[inputLogin].SetValue("root")
	editHostModel.inputs[inputNetworkPort].SetValue("2222")
	editHostModel.inputs[inputIdentityFile].SetValue("id_rsa")

	// Should fail because mandatory fields are not set
	model, messageSequence := editHostModel.save(nil)

	require.Nil(t, messageSequence)
	require.Contains(t, model.title, "not valid")

	// model, messageSequence := editHostModel.save(nil)

	editHostModel.inputs[inputTitle].SetValue("test")
	editHostModel.inputs[inputAddress].SetValue("localhost")

	model, messageSequence = editHostModel.save(nil)

	require.NotNil(t, messageSequence)

	/*
		// Cannot test return values because the function returns an array-like structure of private objects
		expectedSequence := tea.Sequence(
			message.TeaCmd(MsgClose{}),
			message.TeaCmd(message.HostListSelectItem{HostID: model.host.ID}),
			message.TeaCmd(hostlist.MsgRepoUpdated{}),
		)

		one := expectedSequence()
		two := messageSequence() // returns []tea.sequenceMsg, which is private. Cannot test.

		require.Equal(t, one, two)
	*/
}

func Test_Propagate_Title_Value_To_Hostname(t *testing.T) {
	// Test copy values from title to hostname when create a new record in hosts database
	state := state.ApplicationState{}

	storageHostNoFound := mock.NewMockStorage(true)
	editHostModel := New(context.TODO(), storageHostNoFound, &state, &mock.MockLogger{})
	// Check that selected input is title
	assert.Equal(t, editHostModel.focusedInput, inputTitle)

	// Type word 'test' in title
	editHostModel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}})
	editHostModel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	editHostModel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	editHostModel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}})

	// Check that both inputs contain the same values
	require.Equal(t, "test", editHostModel.inputs[inputTitle].Value())
	require.Equal(t, "test", editHostModel.inputs[inputAddress].Value())

	// Select address input
	tmp, _ := editHostModel.Update(tea.KeyMsg{Type: tea.KeyDown})
	editHostModel = tmp.(editModel) // just casting
	// Check that selected input is now address
	assert.Equal(t, editHostModel.focusedInput, inputAddress)

	// Append word 'test' to address so it will become "testtest"
	editHostModel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}})
	editHostModel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	editHostModel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	editHostModel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}})

	// Check that address was updated, but title still preserves the initial value
	require.Equal(t, "test", editHostModel.inputs[inputTitle].Value())
	require.Equal(t, "testtest", editHostModel.inputs[inputAddress].Value())

	// Select title again
	tmp, _ = editHostModel.Update(tea.KeyMsg{Type: tea.KeyUp})
	editHostModel = tmp.(editModel) // just casting
	// Check that selected input is title
	assert.Equal(t, editHostModel.focusedInput, inputTitle)

	// Append '123' to title so it will become "test123"
	editHostModel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'1'}})
	editHostModel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
	editHostModel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'3'}})

	// Check that title was updated, but address still preserves the initial value
	require.Equal(t, "test123", editHostModel.inputs[inputTitle].Value())
	require.Equal(t, "testtest", editHostModel.inputs[inputAddress].Value())
}
