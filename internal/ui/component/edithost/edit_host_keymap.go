package edithost

import (
	"github.com/charmbracelet/bubbles/key"
)

type keyMap struct {
	Up            key.Binding
	Down          key.Binding
	Save          key.Binding
	CopyToAddress key.Binding
	Discard       key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Save, k.CopyToAddress, k.Discard}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return nil
}

var keys = keyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "shift+tab"),
		key.WithHelp("↑", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "tab", "enter"),
		key.WithHelp("↓", "down"),
	),
	Save: key.NewBinding(
		key.WithKeys("ctrl+s"),
		key.WithHelp("ctrl+s", "save"),
	),
	CopyToAddress: key.NewBinding(
		key.WithKeys("alt+enter"),
		key.WithHelp("alt+enter", "copy to address"),
	),
	Discard: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "discard"),
	),
}
