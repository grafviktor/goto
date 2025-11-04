package theme

// DefaultTheme returns the default application theme.
// See cheat-sheet: https://www.ditig.com/publications/256-colors-cheat-sheet
func DefaultTheme() *Theme {
	theme := &Theme{
		Name:        "default",
		Description: "Bubbletea project color palette",
		Colors: ColorsList{
			TextColor:            AdaptiveColor{Light: "#1a1a1a", Dark: "#dddddd"},
			TextColorSelected1:   AdaptiveColor{Light: "#EE6FF8", Dark: "#EE6FF8"},
			TextColorSelected2:   AdaptiveColor{Light: "#F793FF", Dark: "#AD58B4"},
			TextColorError:       AdaptiveColor{Light: "#FF7783", Dark: "#FF7783"},
			TextColorReadonly:    AdaptiveColor{Light: "#585858", Dark: "#585858"},
			TextColorTitle:       AdaptiveColor{Light: "#1a1a1a", Dark: "#dddddd"},
			BackgroundColorTitle: AdaptiveColor{Light: "#5f5fd7", Dark: "#5f5fd7"},
		},
	}

	theme.computeStyles()
	return theme
}
