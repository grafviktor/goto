package input

import "github.com/charmbracelet/lipgloss"

// Ansi to hex color cheat-sheet: https://www.ditig.com/publications/256-colors-cheat-sheet

var (
	focusedStyle = lipgloss.NewStyle().
			BorderForeground(lipgloss.AdaptiveColor{Light: "#F793FF", Dark: "#AD58B4"}).
			Foreground(lipgloss.AdaptiveColor{Light: "#EE6FF8", Dark: "#EE6FF8"})

	errorStyle = lipgloss.NewStyle().
			BorderForeground(lipgloss.AdaptiveColor{Light: "#F793FF", Dark: "#AD58B4"}).
			Foreground(lipgloss.AdaptiveColor{Light: "#FF7783", Dark: "#FF7783"})

	focusedInputText = lipgloss.NewStyle().Foreground(lipgloss.Color("#AD58B4"))
	noStyle          = lipgloss.NewStyle()
	disabledStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#585858"))
)
