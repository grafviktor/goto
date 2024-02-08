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

func TestNotEmptyValidator(t *testing.T) {
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

func TestNetworkPortValidator(t *testing.T) {
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

func TestGetKeyMap(t *testing.T) {
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

func TestSave(t *testing.T) {
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

func TestCopyInputValueFromTo(t *testing.T) {
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

	// Append word 'test' to address, so it will become "testtest"
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

	// Append '123' to title, so it will become "test123"
	editHostModel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'1'}})
	editHostModel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
	editHostModel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'3'}})

	// Check that title was updated, but address still preserves the initial value
	require.Equal(t, "test123", editHostModel.inputs[inputTitle].Value())
	require.Equal(t, "testtest", editHostModel.inputs[inputAddress].Value())
}

func TestHandleCopyInputValueShortcut(t *testing.T) {
	// Test that we can copy values from Title to Address and vice-versa
	// using keyMap.CopyInputValue keyboard shortcut
	state := state.ApplicationState{}
	// That is important in this test. We should make sure that the host which we edit exists
	// in the storage. Otherwise, everything what we type in title will automatically be
	// propagated to address field.
	storageShouldFail := false
	model := New(context.TODO(), mock.NewMockStorage(storageShouldFail), &state, &mock.MockLogger{})
	// Override mock values which we received from mock database and set fields values to 'test'
	model.inputs[inputTitle].SetValue("test")
	model.inputs[inputAddress].SetValue("test")
	// Ensure that selected input is 'Title'
	assert.Equal(t, inputTitle, model.focusedInput)
	// Confirm that 'Title' and 'Host' values are empty strings
	assert.Equal(t, "test", model.inputs[inputTitle].Value())
	assert.Equal(t, "test", model.inputs[inputAddress].Value())
	// Append '123' to title, so it will become 'test123'
	model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'1'}})
	model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
	model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'3'}})
	// Confirm that 'Title' value is 'test123' and 'Address' hasn't changed
	require.Equal(t, "test123", model.inputs[inputTitle].Value())
	require.Equal(t, "test", model.inputs[inputAddress].Value())
	// Now press the shortcut which will copy Title value to Address
	model.Update(tea.KeyMsg{
		Type: tea.KeyEnter,
		Alt:  true,
	})
	// Confirm that 'Title' and 'Host' values are now equal top 'test123'
	assert.Equal(t, "test123", model.inputs[inputTitle].Value())
	assert.Equal(t, "test123", model.inputs[inputAddress].Value())

	// Then select address input and append '456', so the value will be test123456
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyDown})
	assert.Equal(t, inputAddress, updated.(editModel).focusedInput)
	updated.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'4'}})
	updated.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'5'}})
	updated.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'6'}})
	// Check that title still preserves value 'teste123' and address was updated
	assert.Equal(t, "test123", updated.(editModel).inputs[inputTitle].Value())
	assert.Equal(t, "test123456", updated.(editModel).inputs[inputAddress].Value())
	// Now press the shortcut which will copy Address value to Title
	updated.Update(tea.KeyMsg{
		Type: tea.KeyEnter,
		Alt:  true,
	})
	// Ensure that 'Title' and 'Host' values are now equal top 'test123456'
	assert.Equal(t, "test123456", model.inputs[inputTitle].Value())
	assert.Equal(t, "test123456", model.inputs[inputAddress].Value())
}
