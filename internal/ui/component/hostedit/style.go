package hostedit

import (
	"github.com/charmbracelet/lipgloss"

	"github.com/grafviktor/goto/internal/ui/theme"
)

// GetStyles returns the current host edit styles from the theme
func GetStyles() *hostEditStyles {
	t := theme.GetTheme()
	return &hostEditStyles{
		Doc:    t.Styles.HostEdit.Doc,
		Cursor: t.Styles.HostEdit.Cursor,
		Title:  t.Styles.HostEdit.Title,
		Menu:   t.Styles.HostEdit.Menu,
	}
}

type hostEditStyles struct {
	Doc    lipgloss.Style
	Cursor lipgloss.Style
	Title  lipgloss.Style
	Menu   lipgloss.Style
}

// Legacy style variables for backward compatibility
var (
	docStyle    = GetStyles().Doc
	cursorStyle = GetStyles().Cursor
	titleStyle  = GetStyles().Title
	menuStyle   = GetStyles().Menu
)
