package message

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type (
	InitComplete struct{}
	TickMsg      struct{}
)

func Tick() tea.Msg {
	time.Sleep(time.Second * 4)
	return TickMsg{}
}

// A helper function which returns create tea.Cmd from tea.Msg object
func TeaCmd(msg any) func() tea.Msg {
	return func() tea.Msg {
		return msg
	}
}
