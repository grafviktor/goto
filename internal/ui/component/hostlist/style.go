package hostlist

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"

	"github.com/grafviktor/goto/internal/ui/theme"
)

type styles struct {
	cursor       lipgloss.Style
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

	// Group hint - where we display a host group abbreviation in host list title.
	groupAbbreviation lipgloss.Style

	// Status empty - where we display a host group, if it was changed.
	groupHint lipgloss.Style

	// Margings for the whole UI component.
	componentMargins lipgloss.Style
}

func defaultStyles() styles {
	themeSettings := theme.Get().Styles

	return styles{
		componentMargins:     lipgloss.NewStyle().Margin(1, 2, 1, 0), //nolint:mnd // magic numbers are OK for styles
		cursor:               themeSettings.ListExtra.Cursor,
		filterInput:          themeSettings.ListExtra.FilterInput,
		groupAbbreviation:    themeSettings.ListExtra.GroupAbbreviation,
		groupHint:            themeSettings.ListExtra.GroupHint,
		help:                 themeSettings.ListHelp,
		list:                 themeSettings.List,
		listDelegate:         themeSettings.ListDelegate,
		listExtra:            themeSettings.ListExtra,
		paginatorActiveDot:   themeSettings.ListExtra.PaginatorActiveDot,
		paginatorInactiveDot: themeSettings.ListExtra.PaginatorInactiveDot,
		prompt:               themeSettings.ListExtra.Prompt,
	}
}
