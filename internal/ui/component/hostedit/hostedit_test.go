// Package hostedit contains UI components for editing host model attributes.
package hostedit

import (
	"context"
	"fmt"
	"testing"

	"github.com/grafviktor/goto/internal/model/ssh"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/grafviktor/goto/internal/state"
	"github.com/grafviktor/goto/internal/test"
	"github.com/grafviktor/goto/internal/ui/component/hostlist"
	"github.com/grafviktor/goto/internal/ui/message"
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

func TestSave(t *testing.T) {
	hostEditModel := New(context.TODO(), test.NewMockStorage(true), MockAppState(), &test.MockLogger{})
	require.Equal(t, inputTitle, hostEditModel.focusedInput)

	hostEditModel.inputs[inputDescription].SetValue("test")
	hostEditModel.inputs[inputLogin].SetValue("root")
	hostEditModel.inputs[inputNetworkPort].SetValue("2222")
	hostEditModel.inputs[inputIdentityFile].SetValue("id_rsa")

	// Should fail because mandatory fields are not set
	messageSequence := hostEditModel.save(nil)

	require.Nil(t, messageSequence)
	require.Contains(t, hostEditModel.title, "not valid")

	hostEditModel.inputs[inputTitle].SetValue("test")
	hostEditModel.inputs[inputAddress].SetValue("localhost")

	messageSequence = hostEditModel.save(nil)

	require.NotNil(t, messageSequence)

	var dst []tea.Msg
	test.CmdToMessage(messageSequence, &dst)
	require.Contains(t, dst, CloseEditForm{})
	require.Contains(t, dst, hostlist.MsgRefreshRepo{})
	require.Contains(t, dst, message.HostListSelectItem{HostID: 0})
}

func TestCopyInputValueFromTo(t *testing.T) {
	// Test copy values from title to hostname when create a new record in hosts database
	storageHostNoFound := test.NewMockStorage(true)
	hostEditModel := New(context.TODO(), storageHostNoFound, MockAppState(), &test.MockLogger{})
	// Check that selected input is title
	assert.Equal(t, hostEditModel.focusedInput, inputTitle)

	// Type word 'test' in title
	hostEditModel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}})
	hostEditModel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	hostEditModel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	hostEditModel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}})

	// Check that both inputs contain the same values
	require.Equal(t, "test", hostEditModel.inputs[inputTitle].Value())
	require.Equal(t, "test", hostEditModel.inputs[inputAddress].Value())

	// Select address input
	hostEditModel.Update(tea.KeyMsg{Type: tea.KeyDown})
	// Check that selected input is now address
	assert.Equal(t, hostEditModel.focusedInput, inputAddress)

	// Append word 'test' to address, so it will become "testtest"
	hostEditModel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}})
	hostEditModel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	hostEditModel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	hostEditModel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}})

	// Check that address was updated, but title still preserves the initial value
	require.Equal(t, "test", hostEditModel.inputs[inputTitle].Value())
	require.Equal(t, "testtest", hostEditModel.inputs[inputAddress].Value())

	// Select title again
	hostEditModel.Update(tea.KeyMsg{Type: tea.KeyUp})
	// Check that selected input is title
	assert.Equal(t, hostEditModel.focusedInput, inputTitle)

	// Append '123' to title, so it will become "test123"
	hostEditModel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'1'}})
	hostEditModel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
	hostEditModel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'3'}})

	// Check that title was updated, but address still preserves the initial value
	require.Equal(t, "test123", hostEditModel.inputs[inputTitle].Value())
	require.Equal(t, "testtest", hostEditModel.inputs[inputAddress].Value())
}

func TestHandleCopyInputValueShortcut(t *testing.T) {
	// Test that we can copy values from Title to Address and vice-versa
	// using keyMap.CopyInputValue keyboard shortcut

	// That is important in this test - we should make sure that the host which we edit exists
	// in the storage. Otherwise, everything what we type in title will automatically be
	// propagated to address field.
	storageShouldFail := false
	model := New(context.TODO(), test.NewMockStorage(storageShouldFail), MockAppState(), &test.MockLogger{})
	// Override mock values which we received from mock database and set model values to 'test'
	model.host.Title = "test"
	model.host.Address = "test"
	// Update input fields to reflect the model's values
	model.updateInputFields()
	// Ensure that selected input is 'Title'
	assert.Equal(t, inputTitle, model.focusedInput)
	// Confirm that 'Title' and 'Host' values are equal to 'test'
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
	model.Update(tea.KeyMsg{Type: tea.KeyDown})
	assert.Equal(t, inputAddress, model.focusedInput)
	model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'4'}})
	model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'5'}})
	model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'6'}})
	// Check that title still preserves value 'teste123' and address was updated
	assert.Equal(t, "test123", model.inputs[inputTitle].Value())
	assert.Equal(t, "test123456", model.inputs[inputAddress].Value())
	// Now press the shortcut which will copy Address value to Title
	model.Update(tea.KeyMsg{
		Type: tea.KeyEnter,
		Alt:  true,
	})
	// Ensure that 'Title' and 'Host' values are now equal top 'test123456'
	assert.Equal(t, "test123456", model.inputs[inputTitle].Value())
	assert.Equal(t, "test123456", model.inputs[inputAddress].Value())
}

func TestUpdate_TeaSizeMsg(t *testing.T) {
	// Test that if model is ready, WindowSizeMsg message will update viewport
	model := New(context.TODO(), test.NewMockStorage(false), MockAppState(), &test.MockLogger{})
	model.ready = true
	model.Update(tea.WindowSizeMsg{Width: 100, Height: 100})

	require.Greater(t, model.viewport.Height, 0)
	require.Greater(t, model.viewport.Width, 0)
}

func TestView(t *testing.T) {
	// Test that by calling View() function first time, we set ready flag to true
	// and view() returns non-empty string which will be used to build terminal user interface
	model := New(context.TODO(), test.NewMockStorage(false), MockAppState(), &test.MockLogger{})
	assert.False(t, model.ready)
	var ui string = model.View()

	require.True(t, model.ready)
	require.NotEmpty(t, ui)
}

func TestHelpView(t *testing.T) {
	// Test that help view is not empty
	model := New(context.TODO(), test.NewMockStorage(false), MockAppState(), &test.MockLogger{})
	require.NotEmpty(t, model.helpView())
}

func TestHeaderView(t *testing.T) {
	// Test that header view is not empty
	model := New(context.TODO(), test.NewMockStorage(false), MockAppState(), &test.MockLogger{})
	require.NotEmpty(t, model.headerView())
}

func TestHandleDebounceMessage(t *testing.T) {
	// Test that only last message is executed when wrap message in the debounce container
	model := New(context.TODO(), test.NewMockStorage(false), MockAppState(), &test.MockLogger{})
	_, returned1 := model.Update(debouncedMessage{
		wrappedMsg:  struct{}{},
		debounceTag: 0,
	})

	_, returned2 := model.Update(debouncedMessage{
		wrappedMsg:  struct{}{},
		debounceTag: 1,
	})

	_, returned3 := model.Update(debouncedMessage{
		wrappedMsg:  struct{}{},
		debounceTag: 2,
	})

	result1 := returned1()
	result2 := returned2()
	result3 := returned3()

	require.Nil(t, result1)
	require.Nil(t, result2)
	require.NotNil(t, result3)
}

func TestUpdateInputPlaceHolders(t *testing.T) {
	// Make sure that placeholders have correct values once ssh config is changed.
	appState := MockAppState()
	model := New(context.TODO(), test.NewMockStorage(false), appState, &test.MockLogger{})
	model.host.SSHClientConfig = &ssh.Config{
		IdentityFile: "Mock Identity File",
		User:         "Mock User",
		Port:         "Mock Port",
	}
	model.updateInputFields()

	defaultPlaceholderPrefix := "default:"
	require.Equal(t, fmt.Sprintf(
		"%s %s",
		defaultPlaceholderPrefix,
		"Mock User",
	), model.inputs[inputLogin].Placeholder)

	require.Equal(t, fmt.Sprintf(
		"%s %s",
		defaultPlaceholderPrefix,
		"Mock Port",
	), model.inputs[inputNetworkPort].Placeholder)

	require.Equal(t, fmt.Sprintf(
		"%s %s",
		defaultPlaceholderPrefix,
		"Mock Identity File",
	), model.inputs[inputIdentityFile].Placeholder)

	// Now use custom connection settings and make sure that ssh parameters input fields
	// are disabled and placeholders are prefixed with 'readonly:' keyword.
	model.host.Address = "localhost -l root -p 9999 -i ~/.id_rsa"
	model.updateInputFields()

	defaultPlaceholderPrefix = "readonly:"
	require.Equal(t, fmt.Sprintf(
		"%s %s",
		defaultPlaceholderPrefix,
		"Mock User",
	), model.inputs[inputLogin].Placeholder)

	require.Equal(t, fmt.Sprintf(
		"%s %s",
		defaultPlaceholderPrefix,
		"Mock Port",
	), model.inputs[inputNetworkPort].Placeholder)

	require.Equal(t, fmt.Sprintf(
		"%s %s",
		defaultPlaceholderPrefix,
		"Mock Identity File",
	), model.inputs[inputIdentityFile].Placeholder)
}

func MockAppState() *state.ApplicationState {
	return &state.ApplicationState{}
}
