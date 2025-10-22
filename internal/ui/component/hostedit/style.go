package hostedit

import (
	"github.com/charmbracelet/lipgloss"

	"github.com/grafviktor/goto/internal/ui/theme"
)

var (
	themeSettings = theme.GetTheme()

	styleComponentMargins = lipgloss.NewStyle().Margin(1, 0)
	styleSelectedTitle    = themeSettings.Styles.EditForm.SelectedTitle
	styleTitle            = themeSettings.Styles.EditForm.Title
	styleTextReadonly     = themeSettings.Styles.EditForm.TextReadonly
	styleHelp             = themeSettings.Styles.ListHelp
)
