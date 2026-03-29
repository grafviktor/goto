package grouplist

import (
	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/list"
	"charm.land/lipgloss/v2"

	"github.com/grafviktor/goto/internal/ui/theme"
)

type styles struct {
	list         list.Styles
	help         help.Styles
	listDelegate list.DefaultItemStyles
	listExtra    theme.ListExtraStyles

	// Filter styles.
	prompt      lipgloss.Style
	filterInput lipgloss.Style

	// Paginator styles.
	paginatorActiveDot   string
	paginatorInactiveDot string

	// Margins for the whole UI component.
	componentMargins lipgloss.Style
}

func defaultStyles() styles {
	themeSettings := theme.Get().Styles

	return styles{
		componentMargins:     lipgloss.NewStyle().Margin(1, 2, 1, 0), //nolint:mnd // magic nums are OK for styles
		filterInput:          themeSettings.ListExtra.FilterInput,
		help:                 themeSettings.ListHelp,
		list:                 themeSettings.List,
		listDelegate:         themeSettings.ListDelegate,
		listExtra:            themeSettings.ListExtra,
		paginatorActiveDot:   themeSettings.ListExtra.PaginatorActiveDot,
		paginatorInactiveDot: themeSettings.ListExtra.PaginatorInactiveDot,
		prompt:               themeSettings.ListExtra.Prompt,
	}
}
