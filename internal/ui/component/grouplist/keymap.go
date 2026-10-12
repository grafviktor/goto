package grouplist

import "charm.land/bubbles/v2/key"

type keyMap struct {
	selectGroup   key.Binding
	deselectGroup key.Binding
	quitApp       key.Binding
}

func newKeyMap() *keyMap {
	return &keyMap{
		selectGroup: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("↩", "select"),
		),
		deselectGroup: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "deselect"),
		),
		quitApp: key.NewBinding(
			key.WithKeys("q"),
			key.WithHelp("q", "quit"),
		),
	}
}

func (k *keyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		k.selectGroup,
		k.deselectGroup,
		k.quitApp,
	}
}

func (k *keyMap) FullHelp() []key.Binding {
	return []key.Binding{
		k.selectGroup,
		k.deselectGroup,
		k.quitApp,
	}
}
