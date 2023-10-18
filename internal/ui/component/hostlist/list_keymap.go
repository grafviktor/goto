package hostlist

import (
	"github.com/charmbracelet/bubbles/key"
)

type keyMap struct {
	cursorUp   key.Binding
	cursorDown key.Binding
	connect    key.Binding
	append     key.Binding
	clone      key.Binding
	edit       key.Binding
	remove     key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		k.connect,
		k.append,
		k.clone,
		k.edit,
		k.remove,
	}
}

func (k keyMap) FullHelp() []key.Binding {
	return []key.Binding{
		k.connect,
		k.append,
		k.clone,
		k.edit,
		k.remove,
	}
}

/* func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{
			k.connect,
			k.append,
			k.clone,
			k.edit,
			k.remove,
		},
	}
} */

func newDelegateKeyMap() *keyMap {
	return &keyMap{
		cursorUp: key.NewBinding(
			key.WithKeys("up", "shift+tab"),
			key.WithHelp("↑", "up"),
		),
		cursorDown: key.NewBinding(
			key.WithKeys("down", "tab"),
			key.WithHelp("↓", "down"),
		),
		connect: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "connect"),
		),
		append: key.NewBinding(
			key.WithKeys("n"),
			key.WithHelp("n", "new"),
		),
		edit: key.NewBinding(
			key.WithKeys("e"),
			key.WithHelp("e", "edit"),
		),
		clone: key.NewBinding(
			key.WithKeys("c"),
			key.WithHelp("c", "clone"),
		),
		remove: key.NewBinding(
			key.WithKeys("x"),
			key.WithHelp("x", "delete"),
		),
	}
}
