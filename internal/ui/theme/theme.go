// Package theme provides centralized styling and theming functionality
package theme

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
)

// ColorsList defines all colors which can be overridden in the application.
type ColorsList struct {
	BackgroundTitle    AdaptiveColor `json:"backgroundTitle"`
	TextColor          AdaptiveColor `json:"textColor"`
	TextColorError     AdaptiveColor `json:"textColorError"`
	TextColorReadonly  AdaptiveColor `json:"textColorReadonly"`
	TextColorSelected1 AdaptiveColor `json:"textColorSelected1"`
	TextColorSelected2 AdaptiveColor `json:"textColorSelected2"`
}

// AdaptiveColor supports both light and dark theme variants.
type AdaptiveColor struct {
	Light string `json:"light"`
	Dark  string `json:"dark"`
}

// ToLipgloss converts AdaptiveColor to lipgloss.AdaptiveColor.
func (c AdaptiveColor) ToLipgloss() lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: c.Light, Dark: c.Dark}
}

// Theme defines the color scheme and styling for the application.
type Theme struct {
	Name   string     `json:"name"`
	Colors ColorsList `json:"colors"`
	Styles AppStyles  `json:"-"` // Not serialized, computed from colors
}

// DefaultTheme returns the default application theme.
// See cheat-sheet: https://www.ditig.com/publications/256-colors-cheat-sheet
func DefaultTheme() *Theme {
	theme := &Theme{
		Name: "default",
		Colors: ColorsList{
			TextColor:          AdaptiveColor{Light: "#1a1a1a", Dark: "#dddddd"},
			TextColorSelected1: AdaptiveColor{Light: "#EE6FF8", Dark: "#EE6FF8"},
			TextColorSelected2: AdaptiveColor{Light: "#F793FF", Dark: "#AD58B4"},
			TextColorError:     AdaptiveColor{Light: "#FF7783", Dark: "#FF7783"},
			TextColorReadonly:  AdaptiveColor{Light: "#585858", Dark: "#585858"},
			BackgroundTitle:    AdaptiveColor{Light: "#5f5fd7", Dark: "#5f5fd7"},
		},
	}

	theme.computeStyles()
	return theme
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

func (t *Theme) listStyles() list.Styles {
	s := list.DefaultStyles()
	s.TitleBar = lipgloss.NewStyle().Padding(0, 0, 1, 2)
	s.Title = lipgloss.NewStyle().
		Background(t.Colors.BackgroundTitle.ToLipgloss()).
		Foreground(t.Colors.TextColor.ToLipgloss()).
		Padding(0, 1)

	s.StatusBar = lipgloss.NewStyle().
		// Foreground(t.Colors.Muted.ToLipgloss()). // Decided to get rid of specific umatched filtered items color
		Foreground(t.Colors.TextColorReadonly.ToLipgloss()).
		Padding(0, 0, 1, 2)

	s.StatusEmpty = lipgloss.NewStyle().Foreground(t.Colors.TextColorReadonly.ToLipgloss())

	s.StatusBarActiveFilter = lipgloss.NewStyle().
		Foreground(t.Colors.TextColor.ToLipgloss())

	s.StatusBarFilterCount = lipgloss.NewStyle().Foreground(t.Colors.TextColorReadonly.ToLipgloss())

	s.NoItems = lipgloss.NewStyle().
		Foreground(t.Colors.TextColorReadonly.ToLipgloss())

	s.ArabicPagination = lipgloss.NewStyle().Foreground(t.Colors.TextColorReadonly.ToLipgloss())

	s.ActivePaginationDot = s.ActivePaginationDot.Foreground(t.Colors.TextColorSelected1.ToLipgloss())
	s.InactivePaginationDot = s.InactivePaginationDot.Foreground(t.Colors.TextColorReadonly.ToLipgloss())
	s.DividerDot = s.DividerDot.Foreground(t.Colors.TextColorReadonly.ToLipgloss())

	return s
}

func (t *Theme) listDelegateStyles() (s list.DefaultItemStyles) {
	s.NormalTitle = lipgloss.NewStyle().
		Foreground(t.Colors.TextColor.ToLipgloss()).
		Padding(0, 0, 0, 2)

	s.NormalDesc = s.NormalTitle.
		Foreground(t.Colors.TextColorReadonly.ToLipgloss())

	s.SelectedTitle = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(t.Colors.TextColorSelected2.ToLipgloss()).
		Foreground(t.Colors.TextColorSelected1.ToLipgloss()).
		Padding(0, 0, 0, 1)

	s.SelectedDesc = s.SelectedTitle.
		Foreground(t.Colors.TextColorSelected2.ToLipgloss())

	s.DimmedTitle = lipgloss.NewStyle().
		Foreground(t.Colors.TextColorReadonly.ToLipgloss()).
		Padding(0, 0, 0, 2)

	s.DimmedDesc = s.DimmedTitle.
		Foreground(t.Colors.TextColorReadonly.ToLipgloss())

	s.FilterMatch = lipgloss.NewStyle().Underline(true)

	return s
}

func (t *Theme) listHelpStyles() (s help.Styles) {
	s.ShortKey = s.ShortKey.Foreground(t.Colors.TextColorReadonly.ToLipgloss())
	s.ShortDesc = s.ShortKey.Foreground(t.Colors.TextColorReadonly.ToLipgloss())
	s.ShortSeparator = s.ShortKey.Foreground(t.Colors.TextColorReadonly.ToLipgloss())

	s.FullKey = s.FullKey.Foreground(t.Colors.TextColorReadonly.ToLipgloss())
	s.FullDesc = s.FullKey.Foreground(t.Colors.TextColorReadonly.ToLipgloss())
	s.FullSeparator = s.FullKey.Foreground(t.Colors.TextColorReadonly.ToLipgloss())

	return s
}

type ListExtraStyles struct {
	GroupAbbreviation lipgloss.Style
	GroupHint         lipgloss.Style
	Prompt            lipgloss.Style
	FilterInput       lipgloss.Style
	// These 2 paginator styles are string values.
	PaginatorActiveDot   string
	PaginatorInactiveDot string
}

func (t *Theme) listExtraStyles() (s ListExtraStyles) {
	s.GroupAbbreviation = lipgloss.NewStyle().
		Background(t.Colors.TextColor.ToLipgloss()).
		Foreground(t.Colors.BackgroundTitle.ToLipgloss()).
		Padding(0, 1)

	s.GroupHint = lipgloss.NewStyle().
		Foreground(t.Colors.TextColorReadonly.ToLipgloss())

	s.Prompt = lipgloss.NewStyle().
		Foreground(t.Colors.TextColorSelected1.ToLipgloss())
	s.FilterInput = lipgloss.NewStyle().
		Foreground(t.Colors.TextColor.ToLipgloss())

	s.PaginatorActiveDot = lipgloss.NewStyle().
		Foreground(t.Colors.TextColor.ToLipgloss()).
		SetString("•").String()
	s.PaginatorInactiveDot = lipgloss.NewStyle().
		Foreground(t.Colors.TextColorReadonly.ToLipgloss()).
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

func (t *Theme) inputStyles() (s InputStyles) {
	s.InputFocused = lipgloss.NewStyle().
		BorderForeground(t.Colors.TextColorSelected2.ToLipgloss()).
		Foreground(t.Colors.TextColorSelected1.ToLipgloss())
	s.InputError = lipgloss.NewStyle().
		BorderForeground(t.Colors.TextColorSelected2.ToLipgloss()).
		Foreground(t.Colors.TextColorError.ToLipgloss())
	s.TextFocused = lipgloss.NewStyle().
		Foreground(t.Colors.TextColorSelected2.ToLipgloss())
	s.TextReadonly = lipgloss.NewStyle().
		Foreground(t.Colors.TextColorReadonly.ToLipgloss())
	s.TextNormal = lipgloss.NewStyle().
		Foreground(t.Colors.TextColor.ToLipgloss())

	return s
}

// EditForm contains styles for host edit components.
type EditForm struct {
	VericalMargin lipgloss.Style
	SelectedTitle lipgloss.Style
	Title         lipgloss.Style
	TextReadonly  lipgloss.Style
}

func (t *Theme) editFormStyles() (s EditForm) {
	s.VericalMargin = lipgloss.NewStyle().Margin(1, 0)
	s.SelectedTitle = lipgloss.NewStyle().
		BorderForeground(t.Colors.TextColorSelected2.ToLipgloss()).
		Foreground(t.Colors.TextColorSelected1.ToLipgloss())
	s.Title = lipgloss.NewStyle().
		Background(t.Colors.BackgroundTitle.ToLipgloss()).
		Foreground(t.Colors.TextColor.ToLipgloss()).
		Padding(0, 1).
		Margin(1, 2, 0)
	s.TextReadonly = lipgloss.NewStyle().
		Foreground(t.Colors.TextColorReadonly.ToLipgloss()).
		Margin(2, 2, 1)
	return s
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

// LoadThemeFromFile loads a theme from a JSON file.
func LoadThemeFromFile(filePath string) (*Theme, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read theme file: %w", err)
	}

	var theme Theme
	if err := json.Unmarshal(data, &theme); err != nil {
		return nil, fmt.Errorf("failed to parse theme file: %w", err)
	}

	theme.computeStyles()
	return &theme, nil
}

// SaveThemeToFile saves a theme to a JSON file.
func SaveThemeToFile(theme *Theme, filePath string) error {
	// Ensure directory exists
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	data, err := json.MarshalIndent(theme, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal theme: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write theme file: %w", err)
	}

	return nil
}

var currentTheme *Theme

// SetTheme sets the current application theme.
func SetTheme(theme *Theme) {
	currentTheme = theme
}

// GetTheme returns the current application theme.
func GetTheme() *Theme {
	if currentTheme == nil {
		currentTheme = DefaultTheme()
	}
	return currentTheme
}

// LoadTheme loads a theme from file or falls back to default.
func LoadTheme(configDir string) *Theme {
	themeFile := filepath.Join(configDir, "theme.json")

	theme, err := LoadThemeFromFile(themeFile)
	if err != nil {
		// Fall back to default theme
		theme = DefaultTheme()
	}

	SetTheme(theme)
	return theme
}
