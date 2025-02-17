package hostlist

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
)

type Styles struct {
	list.Styles
	Group lipgloss.Style
	Title lipgloss.Style
}

func CustomStyles() Styles {
	styles := list.DefaultStyles()
	styles.Title = lipgloss.NewStyle()

	return Styles{
		Styles: styles,
		Group: lipgloss.NewStyle().
			Background(lipgloss.Color("#ffffd7")).
			Foreground(lipgloss.Color("#5f5fd7")).
			Padding(0, 1),
		Title: lipgloss.NewStyle().
			Background(lipgloss.Color("62")).
			Foreground(lipgloss.Color("230")).
			Padding(0, 1),
	}
}
