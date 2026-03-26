package input

import (
	"charm.land/lipgloss/v2"

	"github.com/grafviktor/goto/internal/ui/theme"
)

type styles struct {
	cursor       lipgloss.Style
	inputError   lipgloss.Style
	inputFocused lipgloss.Style
	textNormal   lipgloss.Style
	textFocused  lipgloss.Style
	textReadonly lipgloss.Style
}

func defaultStyles() styles {
	themeSettings := theme.Get().Styles.Input

	return styles{
		cursor:       themeSettings.Cursor,
		inputError:   themeSettings.InputError,
		inputFocused: themeSettings.InputFocused,
		textFocused:  themeSettings.TextFocused,
		textNormal:   themeSettings.TextNormal,
		textReadonly: themeSettings.TextReadonly,
	}
}
