package hostlist

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/samber/lo"
)

type keyMapStateEnum string

var keyMapState = struct {
	EditkeysHidden         keyMapStateEnum
	EditkeysPartiallyShown keyMapStateEnum
	EditkeysShown          keyMapStateEnum
}{
	EditkeysHidden:         "hidden",
	EditkeysPartiallyShown: "partially shown",
	EditkeysShown:          "shown",
}

type keyMap struct {
	cursorUp     key.Binding
	cursorDown   key.Binding
	selectGroup  key.Binding
	connect      key.Binding
	copyID       key.Binding
	append       key.Binding
	clone        key.Binding
	edit         key.Binding
	remove       key.Binding
	toggleLayout key.Binding
	confirm      key.Binding
	keyMapState  keyMapStateEnum
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
		selectGroup: key.NewBinding(
			key.WithKeys("z"),
			key.WithHelp("z", "group"),
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
			key.WithHelp("t", "ssh-copy-id"),
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

	km.keyMapState = keyMapState.EditkeysHidden
	return &km
}

func (k *keyMap) keysForNullHost() {
	if k.keyMapState != keyMapState.EditkeysHidden {
		k.keyMapState = keyMapState.EditkeysHidden
		k.keysSetEnabled(false)
	}
}

func (k *keyMap) keysForReadonlyHost() {
	if k.keyMapState != keyMapState.EditkeysPartiallyShown {
		k.keyMapState = keyMapState.EditkeysPartiallyShown
		k.clone.SetEnabled(false)
		k.connect.SetEnabled(true)
		k.copyID.SetEnabled(true)
		k.cursorDown.SetEnabled(true)
		k.cursorUp.SetEnabled(true)
		k.edit.SetEnabled(true)
		k.remove.SetEnabled(false)
	}
}

func (k *keyMap) keysForWritableHost() {
	if k.keyMapState != keyMapState.EditkeysShown {
		k.keyMapState = keyMapState.EditkeysShown
		k.keysSetEnabled(true)
	}
}

func (k *keyMap) keysSetEnabled(val bool) {
	k.clone.SetEnabled(val)
	k.connect.SetEnabled(val)
	k.cursorDown.SetEnabled(val)
	k.cursorUp.SetEnabled(val)
	k.edit.SetEnabled(val)
	k.remove.SetEnabled(val)
	k.copyID.SetEnabled(val)
}

func (k *keyMap) UpdateKeyVisibility(item list.Item) string {
	host, ok := item.(ListItemHost)
	if !ok { //nolint:gocritic // it's more readable in if-else, then in switch-case block
		k.keysForNullHost()
	} else if host.IsReadOnly() {
		k.keysForReadonlyHost()
	} else {
		k.keysForWritableHost()
	}

	return string(k.keyMapState)
}

func (k *keyMap) ShortHelp() []key.Binding {
	tmp := []key.Binding{
		k.connect,
		k.append,
		k.clone,
		k.edit,
		k.remove,
		k.selectGroup,
	}

	// Hide all disabled key shortcuts from the screen
	return lo.Filter(tmp, func(k key.Binding, _ int) bool {
		return k.Enabled()
	})
}

func (k *keyMap) FullHelp() []key.Binding {
	return []key.Binding{
		k.connect,
		k.append,
		k.clone,
		k.edit,
		k.remove,
		k.selectGroup,
		k.copyID,
		k.toggleLayout,
	}
}
