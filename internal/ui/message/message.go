package message

import tea "github.com/charmbracelet/bubbletea"

type (
	GlobalInitComplete struct{}
	GlobalCreate       struct{}
)

// A helper function which returns create tea.Cmd from tea.Msg object
func TeaCmd(msg any) func() tea.Msg {
	return func() tea.Msg {
		return msg
	}
}
