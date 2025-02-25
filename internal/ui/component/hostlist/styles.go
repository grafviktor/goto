package hostlist

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
)

type styles struct {
	list.Styles
	Group lipgloss.Style
	Title lipgloss.Style
}

func customStyles() styles {
	defaultStyles := list.DefaultStyles()
	defaultStyles.Title = lipgloss.NewStyle()

	return styles{
		Styles: defaultStyles,
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
