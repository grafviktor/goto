package input

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInput_Update_KeyMsg(t *testing.T) {
	// Test that the input component ignores key messages when it's disabled

	model := New()
	model.Focus()
	model.Update(tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune{'t', 'e', 's', 't'},
	})

	assert.Equal(t, "test", model.Value())
	model.SetValue("")
	model.SetEnabled(false)

	model.Update(tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune{'t', 'e', 's', 't', '2'},
	})
	require.Equal(t, "", model.Value())
}

func TestInput_Display_Tooltip(t *testing.T) {
	// Test that the input component displays a tooltip only when the tooltip is enabled

	model := New()
	model.Focus()
	model.Tooltip = "mock tooltip"
	model.SetValue("mock text")

	model.SetDisplayTooltip(true)
	require.Contains(t, model.View(), "mock tooltip")
	require.Contains(t, model.View(), "mock text")

	model.SetDisplayTooltip(false)
	require.NotContains(t, model.View(), "mock tooltip")
	require.Contains(t, model.View(), "mock text")
}
