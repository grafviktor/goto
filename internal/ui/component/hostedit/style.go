package hostedit

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/lipgloss"

	"github.com/grafviktor/goto/internal/ui/theme"
)

type styles struct {
	componentMargins lipgloss.Style
	cursor           lipgloss.Style
	selectedTitle    lipgloss.Style
	title            lipgloss.Style
	textReadonly     lipgloss.Style
	help             help.Styles
}

func defaultStyles() styles {
	themeSettings := theme.Get().Styles

	return styles{
		componentMargins: lipgloss.NewStyle().Margin(1, 0),
		cursor:           themeSettings.ListExtra.Cursor,
		help:             themeSettings.ListHelp,
		selectedTitle:    themeSettings.EditForm.SelectedTitle,
		textReadonly:     themeSettings.EditForm.TextReadonly,
		title:            themeSettings.EditForm.Title,
	}
}
