// Package theme provides centralized styling and theming functionality
//
//nolint:mnd // Ignore magic numbers in styling: paddings and margins
package theme

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
)

// AdaptiveColor supports both light and dark theme variants.
type AdaptiveColor struct {
	Light string `json:"light"`
	Dark  string `json:"dark"`
}

// toLipgloss converts AdaptiveColor to lipgloss.AdaptiveColor.
func (c AdaptiveColor) toLipgloss() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: c.Light, Dark: c.Dark}
}

// ColorsList defines all colors which can be overridden in the application.
type ColorsList struct {
	TextColor            AdaptiveColor `json:"textColor"`
	TextColorError       AdaptiveColor `json:"textColorError"`
	TextColorReadonly    AdaptiveColor `json:"textColorReadonly"`
	TextColorSelected1   AdaptiveColor `json:"textColorSelected1"`
	TextColorSelected2   AdaptiveColor `json:"textColorSelected2"`
	TextColorTitle       AdaptiveColor `json:"textColorTitle"`
	BackgroundColorTitle AdaptiveColor `json:"backgroundColorTitle"`
}

// Theme defines the color scheme and styling for the application.
type Theme struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Colors      ColorsList `json:"colors"`
	Styles      AppStyles  `json:"-"` // Not serialized, computed from colors
}

// AppStyles contains all computed styles for the application.
type AppStyles struct {
	Input        InputStyles
	List         list.Styles
	ListDelegate list.DefaultItemStyles
	ListHelp     help.Styles
	EditForm     EditForm
	ListExtra    ListExtraStyles
}

func (t *Theme) computeStyles() {
	t.Styles = AppStyles{
		Input:        t.inputStyles(),
		ListExtra:    t.listExtraStyles(),
		List:         t.listStyles(),
		ListDelegate: t.listDelegateStyles(),
		ListHelp:     t.listHelpStyles(),
		EditForm:     t.editFormStyles(),
	}
}

func (t *Theme) listStyles() list.Styles {
	s := list.DefaultStyles()
	s.TitleBar = lipgloss.NewStyle().Padding(0, 0, 1, 2)
	s.Title = lipgloss.NewStyle().
		Background(t.Colors.BackgroundColorTitle.toLipgloss()).
		Foreground(t.Colors.TextColorTitle.toLipgloss()).
		Padding(0, 1)

	s.StatusBar = lipgloss.NewStyle().
		Foreground(t.Colors.TextColorReadonly.toLipgloss()).
		Padding(0, 0, 1, 2)

	s.StatusEmpty = lipgloss.NewStyle().Foreground(t.Colors.TextColorReadonly.toLipgloss())

	s.StatusBarActiveFilter = lipgloss.NewStyle().
		Foreground(t.Colors.TextColor.toLipgloss())

	s.StatusBarFilterCount = lipgloss.NewStyle().Foreground(t.Colors.TextColorReadonly.toLipgloss())

	s.NoItems = lipgloss.NewStyle().
		Foreground(t.Colors.TextColorReadonly.toLipgloss())

	s.ArabicPagination = lipgloss.NewStyle().Foreground(t.Colors.TextColorReadonly.toLipgloss())

	s.ActivePaginationDot = s.ActivePaginationDot.Foreground(t.Colors.TextColorSelected1.toLipgloss())
	s.InactivePaginationDot = s.InactivePaginationDot.Foreground(t.Colors.TextColorReadonly.toLipgloss())
	s.DividerDot = s.DividerDot.Foreground(t.Colors.TextColorReadonly.toLipgloss())

	return s
}

func (t *Theme) listDelegateStyles() list.DefaultItemStyles {
	s := list.DefaultItemStyles{}
	s.NormalTitle = lipgloss.NewStyle().
		Foreground(t.Colors.TextColor.toLipgloss()).
		Padding(0, 0, 0, 2)

	s.NormalDesc = s.NormalTitle.
		Foreground(t.Colors.TextColorReadonly.toLipgloss())

	s.SelectedTitle = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(t.Colors.TextColorSelected2.toLipgloss()).
		Foreground(t.Colors.TextColorSelected1.toLipgloss()).
		Padding(0, 0, 0, 1)

	s.SelectedDesc = s.SelectedTitle.
		Foreground(t.Colors.TextColorSelected2.toLipgloss())

	s.DimmedTitle = lipgloss.NewStyle().
		Foreground(t.Colors.TextColorReadonly.toLipgloss()).
		Padding(0, 0, 0, 2)

	s.DimmedDesc = s.DimmedTitle.
		Foreground(t.Colors.TextColorReadonly.toLipgloss())

	s.FilterMatch = lipgloss.NewStyle().Underline(true)

	return s
}

func (t *Theme) listHelpStyles() help.Styles {
	s := help.Styles{}

	s.ShortKey = s.ShortKey.Foreground(t.Colors.TextColorReadonly.toLipgloss())
	s.ShortDesc = s.ShortKey.Foreground(t.Colors.TextColorReadonly.toLipgloss())
	s.ShortSeparator = s.ShortKey.Foreground(t.Colors.TextColorReadonly.toLipgloss())

	s.FullKey = s.FullKey.Foreground(t.Colors.TextColorReadonly.toLipgloss())
	s.FullDesc = s.FullKey.Foreground(t.Colors.TextColorReadonly.toLipgloss())
	s.FullSeparator = s.FullKey.Foreground(t.Colors.TextColorReadonly.toLipgloss())

	return s
}

type ListExtraStyles struct {
	Cursor            lipgloss.Style
	GroupAbbreviation lipgloss.Style
	GroupHint         lipgloss.Style
	Prompt            lipgloss.Style
	FilterInput       lipgloss.Style
	// These 2 paginator styles are string values.
	PaginatorActiveDot   string
	PaginatorInactiveDot string
}

func (t *Theme) listExtraStyles() ListExtraStyles {
	s := ListExtraStyles{}

	s.Cursor = lipgloss.NewStyle().
		Foreground(t.Colors.TextColorSelected2.toLipgloss())

	s.GroupAbbreviation = lipgloss.NewStyle().
		// Swap colors between each other to separate group abbreviation from title.
		Background(t.Colors.TextColorTitle.toLipgloss()).
		Foreground(t.Colors.BackgroundColorTitle.toLipgloss()).
		Padding(0, 1)

	s.GroupHint = lipgloss.NewStyle().
		Foreground(t.Colors.TextColorReadonly.toLipgloss())

	s.Prompt = lipgloss.NewStyle().
		Foreground(t.Colors.TextColorSelected1.toLipgloss())
	s.FilterInput = lipgloss.NewStyle().
		Foreground(t.Colors.TextColor.toLipgloss())

	// These 2 paginator styles are string values.
	s.PaginatorActiveDot = lipgloss.NewStyle().
		Foreground(t.Colors.TextColor.toLipgloss()).
		SetString("•").String()
	s.PaginatorInactiveDot = lipgloss.NewStyle().
		Foreground(t.Colors.TextColorReadonly.toLipgloss()).
		SetString("•").String()

	return s
}

// InputStyles contains styles for input components.
type InputStyles struct {
	InputFocused lipgloss.Style
	InputError   lipgloss.Style
	TextFocused  lipgloss.Style
	TextReadonly lipgloss.Style
	TextNormal   lipgloss.Style
}

func (t *Theme) inputStyles() InputStyles {
	s := InputStyles{}
	s.InputFocused = lipgloss.NewStyle().
		BorderForeground(t.Colors.TextColorSelected2.toLipgloss()).
		Foreground(t.Colors.TextColorSelected1.toLipgloss())
	s.InputError = lipgloss.NewStyle().
		BorderForeground(t.Colors.TextColorSelected2.toLipgloss()).
		Foreground(t.Colors.TextColorError.toLipgloss())
	s.TextFocused = lipgloss.NewStyle().
		Foreground(t.Colors.TextColorSelected2.toLipgloss())
	s.TextReadonly = lipgloss.NewStyle().
		Foreground(t.Colors.TextColorReadonly.toLipgloss())
	s.TextNormal = lipgloss.NewStyle().
		Foreground(t.Colors.TextColor.toLipgloss())

	return s
}

// EditForm contains styles for host edit components.
type EditForm struct {
	SelectedTitle lipgloss.Style
	Title         lipgloss.Style
	TextReadonly  lipgloss.Style
}

func (t *Theme) editFormStyles() EditForm {
	s := EditForm{}

	s.SelectedTitle = lipgloss.NewStyle().
		BorderForeground(t.Colors.TextColorSelected2.toLipgloss()).
		Foreground(t.Colors.TextColorSelected1.toLipgloss())
	s.Title = lipgloss.NewStyle().
		Background(t.Colors.BackgroundColorTitle.toLipgloss()).
		Foreground(t.Colors.TextColorTitle.toLipgloss()).
		Padding(0, 1).
		Margin(1, 2, 0)
	s.TextReadonly = lipgloss.NewStyle().
		Foreground(t.Colors.TextColorReadonly.toLipgloss()).
		Margin(2, 2, 1)
	return s
}
