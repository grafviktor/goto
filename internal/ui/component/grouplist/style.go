package grouplist

import (
	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/list"
	"charm.land/lipgloss/v2"

	"github.com/grafviktor/goto/internal/ui/theme"
)

type styles struct {
	styleList         list.Styles
	styleHelp         help.Styles
	styleListDelegate list.DefaultItemStyles
	styleListExtra    theme.ListExtraStyles

	// Filter styles.
	stylePrompt      lipgloss.Style
	styleFilterInput lipgloss.Style

	// Paginator styles.
	stylePaginatorActiveDot   string
	stylePaginatorInactiveDot string

	// Margins for the whole UI component.
	styleComponentMargins lipgloss.Style
}

func defaultStyles() styles {
	themeSettings := theme.Get().Styles

	return styles{
		styleComponentMargins:     lipgloss.NewStyle().Margin(1, 2, 1, 0), //nolint:mnd // magic nums are OK for styles
		styleFilterInput:          themeSettings.ListExtra.FilterInput,
		styleHelp:                 themeSettings.ListHelp,
		styleList:                 themeSettings.List,
		styleListDelegate:         themeSettings.ListDelegate,
		styleListExtra:            themeSettings.ListExtra,
		stylePaginatorActiveDot:   themeSettings.ListExtra.PaginatorActiveDot,
		stylePaginatorInactiveDot: themeSettings.ListExtra.PaginatorInactiveDot,
		stylePrompt:               themeSettings.ListExtra.Prompt,
	}
}
