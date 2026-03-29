package hostedit

import (
	"charm.land/bubbles/v2/help"
	"charm.land/lipgloss/v2"

	"github.com/grafviktor/goto/internal/ui/theme"
)

type styles struct {
	componentMargins lipgloss.Style
	keyMap           lipgloss.Style
	selectedTitle    lipgloss.Style
	title            lipgloss.Style
	textReadonly     lipgloss.Style
	help             help.Styles
}

func defaultStyles() styles {
	themeSettings := theme.Get().Styles

	return styles{
		componentMargins: lipgloss.NewStyle().Margin(1, 0),
		help:             themeSettings.ListHelp,
		keyMap:           themeSettings.EditForm.KeyMap,
		selectedTitle:    themeSettings.EditForm.SelectedTitle,
		textReadonly:     themeSettings.EditForm.TextReadonly,
		title:            themeSettings.EditForm.Title,
	}
}
