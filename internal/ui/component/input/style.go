package input

import (
	"github.com/charmbracelet/lipgloss"

	"github.com/grafviktor/goto/internal/ui/theme"
)

type styles struct {
	inputError   lipgloss.Style
	inputFocused lipgloss.Style
	textNormal   lipgloss.Style
	textFocused  lipgloss.Style
	textReadonly lipgloss.Style
}

func defaultStyles() styles {
	themeSettings := theme.GetTheme().Styles.Input

	return styles{
		inputError:   themeSettings.InputError,
		inputFocused: themeSettings.InputFocused,
		textFocused:  themeSettings.TextFocused,
		textNormal:   themeSettings.TextNormal,
		textReadonly: themeSettings.TextReadonly,
	}
}
