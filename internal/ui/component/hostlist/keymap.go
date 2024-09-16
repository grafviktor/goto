package hostlist

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/samber/lo"
)

type keyMap struct {
	cursorUp              key.Binding
	cursorDown            key.Binding
	connect               key.Binding
	copyID                key.Binding
	append                key.Binding
	clone                 key.Binding
	edit                  key.Binding
	remove                key.Binding
	toggleLayout          key.Binding
	confirm               key.Binding
	shouldShowEditButtons bool
}

func newDelegateKeyMap() *keyMap {
	km := keyMap{
		cursorUp: key.NewBinding(
			key.WithKeys("up", "k", "shift+tab"),
			key.WithHelp("↑/k", "up"),
		),
		cursorDown: key.NewBinding(
			key.WithKeys("down", "j", "tab"),
			key.WithHelp("↓/j", "down"),
		),
		connect: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("↩", "connect"),
		),
		append: key.NewBinding(
			key.WithKeys("i", "n", "insert"),
			key.WithHelp("i/n", "new"),
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
			key.WithKeys("d", "x"),
			key.WithHelp("d/x", "delete"),
		),
		copyID: key.NewBinding(
			key.WithKeys("t"),
			key.WithHelp("t", "ssh copy id"),
		),
		toggleLayout: key.NewBinding(
			key.WithKeys("v"),
			key.WithHelp("v", "toggle view"),
		),
		confirm: key.NewBinding(
			key.WithKeys("y", "Y"),
			key.WithHelp("y", "confirm"),
		),
	}

	km.shouldShowEditButtons = true
	return &km
}

func (k *keyMap) SetShouldShowEditButtons(val bool) {
	k.shouldShowEditButtons = val
	k.clone.SetEnabled(val)
	k.connect.SetEnabled(val)
	k.copyID.SetEnabled(val)
	k.cursorDown.SetEnabled(val)
	k.cursorUp.SetEnabled(val)
	k.edit.SetEnabled(val)
	k.remove.SetEnabled(val)
}

func (k *keyMap) ShouldShowEditButtons() bool {
	return k.shouldShowEditButtons
}

func (k *keyMap) ShortHelp() []key.Binding {
	tmp := []key.Binding{
		k.connect,
		k.append,
		k.clone,
		k.edit,
		k.remove,
	}

	// Hide all disabled key shortcuts from the screen
	return lo.Filter(tmp, func(k key.Binding, _ int) bool {
		return k.Enabled()
	})
}

func (k *keyMap) FullHelp() []key.Binding {
	return []key.Binding{
		k.connect,
		k.copyID,
		k.append,
		k.clone,
		k.edit,
		k.remove,
		k.toggleLayout,
	}
}
