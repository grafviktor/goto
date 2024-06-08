package hostlist

import (
	"github.com/charmbracelet/bubbles/key"
)

type keyMap struct {
	cursorUp              key.Binding
	cursorDown            key.Binding
	append                key.Binding
	confirm               key.Binding
	toggleLayout          key.Binding
	shouldShowEditButtons bool
}

func newListKeyMap() *keyMap {
	return &keyMap{
		cursorUp: key.NewBinding(
			key.WithKeys("up", "k", "shift+tab"),
			key.WithHelp("↑/k", "up"),
		),
		cursorDown: key.NewBinding(
			key.WithKeys("down", "j", "tab"),
			key.WithHelp("↓/j", "down"),
		),
		append: key.NewBinding(
			key.WithKeys("n"),
			key.WithHelp("n", "new"),
		),
		confirm: key.NewBinding(
			key.WithKeys("y", "Y"),
			key.WithHelp("y", "confirm"),
		),
		toggleLayout: key.NewBinding(
			key.WithKeys("v"),
			key.WithHelp("v", "view"),
		),
	}
}

func (k *keyMap) SetShouldShowEditButtons(val bool) {
	// k.shouldShowEditButtons = val
	// k.clone.SetEnabled(val)
	// k.connect.SetEnabled(val)
	// k.cursorDown.SetEnabled(val)
	// k.cursorUp.SetEnabled(val)
	// k.edit.SetEnabled(val)
	// k.remove.SetEnabled(val)
}

func (k *keyMap) ShouldShowEditButtons() bool {
	return k.shouldShowEditButtons
}

func (k *keyMap) ShortHelp() []key.Binding {
	// tmp := []key.Binding{
	// 	k.connect,
	// 	k.append,
	// 	k.clone,
	// 	k.edit,
	// 	k.remove,
	// }

	// Hide all disabled key shortcuts from the screen
	// return lo.Filter(tmp, func(k key.Binding, _ int) bool {
	// 	return k.Enabled()
	// })

	return []key.Binding{
		k.append,
		k.toggleLayout,
	}
}

func (k *keyMap) FullHelp() []key.Binding {
	return []key.Binding{
		k.append,
		k.toggleLayout,
		// k.connect,
		// k.append,
		// k.clone,
		// k.edit,
		// k.remove,
		// k.cursorUp,
		// k.cursorDown,
	}
}
