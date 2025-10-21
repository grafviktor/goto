package hostedit

import (
	"github.com/charmbracelet/lipgloss"

	"github.com/grafviktor/goto/internal/ui/theme"
)

// GetStyles returns the current host edit styles from the theme
func GetStyles() *hostEditStyles {
	t := theme.GetTheme()
	return &hostEditStyles{
		VerticalMargin: t.Styles.EditForm.VericalMargin,
		SelectedTitle:  t.Styles.EditForm.SelectedTitle,
		Title:          t.Styles.EditForm.Title,
		Menu:           t.Styles.EditForm.TextReadonly,
	}
}

type hostEditStyles struct {
	VerticalMargin lipgloss.Style
	SelectedTitle  lipgloss.Style
	Title          lipgloss.Style
	Menu           lipgloss.Style
}

// Legacy style variables for backward compatibility.
var (
	docStyle    = GetStyles().VerticalMargin
	cursorStyle = GetStyles().SelectedTitle
	titleStyle  = GetStyles().Title
	menuStyle   = GetStyles().Menu
)
