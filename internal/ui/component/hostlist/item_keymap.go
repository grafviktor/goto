package hostlist

import (
	"github.com/charmbracelet/bubbles/key"
)

type HostDelegateKeyMap struct {
	connect key.Binding
	clone   key.Binding
	edit    key.Binding
	remove  key.Binding
}

func newHostDelegateKeyMap() *HostDelegateKeyMap {
	return &HostDelegateKeyMap{
		connect: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("↩", "connect"),
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
			key.WithKeys("d"),
			key.WithHelp("d", "delete"),
		),
	}
}

func (k *HostDelegateKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		k.connect,
		k.clone,
		k.edit,
		k.remove,
	}
}

func (k *HostDelegateKeyMap) FullHelp() []key.Binding {
	return []key.Binding{
		k.connect,
		k.clone,
		k.edit,
		k.remove,
	}
}
