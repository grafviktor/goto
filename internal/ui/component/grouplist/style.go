package grouplist

import (
	"github.com/charmbracelet/lipgloss"

	"github.com/grafviktor/goto/internal/ui/theme"
)

var (
	themeSettings = theme.GetTheme()

	styleList         = themeSettings.Styles.List
	styleHelp         = themeSettings.Styles.ListHelp
	styleListDelegate = themeSettings.Styles.ListDelegate
	styleListExtra    = themeSettings.Styles.ListExtra
	// Filter styles.
	stylePrompt      = styleListExtra.Prompt
	styleFilterInput = styleListExtra.FilterInput
	// Paginator styles.
	stylePaginatorActiveDot   = styleListExtra.PaginatorActiveDot
	stylePaginatorInactiveDot = styleListExtra.PaginatorInactiveDot
	// Margings for the whole UI component.
	styleComponentMargins = lipgloss.NewStyle().Margin(1, 2, 1, 0) //nolint:mnd // magic numbers are OK fo styles
)
