package input

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInput_Update_KeyMsg(t *testing.T) {
	// Test that the input component ignores key messages when it's disabled

	model := New()
	model.Focus()
	model.Update(tea.KeyPressMsg{
		Text: "test",
	})

	assert.Equal(t, "test", model.Value())
	model.SetValue("")
	model.SetEnabled(false)

	model.Update(tea.KeyPressMsg{
		Text: "test2",
	})
	require.Empty(t, model.Value())
}

func TestInput_Display_Tooltip(t *testing.T) {
	// Test that the input component displays a tooltip only when the tooltip is enabled

	model := New()
	model.Focus()
	model.Tooltip = "mock tooltip"
	model.SetValue("mock text")

	model.SetDisplayTooltip(true)
	require.Contains(t, model.View().Content, "mock tooltip")
	require.Contains(t, model.View().Content, "mock text")

	model.SetDisplayTooltip(false)
	require.NotContains(t, model.View().Content, "mock tooltip")
	require.Contains(t, model.View().Content, "mock text")
}

func Test_setPlaceholderWidth(t *testing.T) {
	// Check that workaround for bubletea placeholder bug works
	// If input contains a value, then its width should be zero, this way the model
	// will calculate the component width automatically.
	// If there is not value, but placehoolder is set, then the width should be equal to the placeholder width.
	// Without this workaround the component will only display the first character of the placeholder.
	model := New()
	model.Focus()
	require.Zero(t, model.Width())
	// Set placeholder, model width must be updated accordingly.
	model.Placeholder = "This is a placeholder"
	model.setPlaceholderWidth()
	require.Equal(t, 21, model.Width())
	// Set value, model width must be reset to zero.
	model.SetValue("some value")
	require.Equal(t, 0, model.Width())
}
