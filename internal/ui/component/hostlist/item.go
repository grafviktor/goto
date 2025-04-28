package hostlist

import (
	"github.com/grafviktor/goto/internal/model/host"
	"github.com/samber/lo"
)

// ListItemHost is an adaptor between host model and bubbletea list model.
type ListItemHost struct {
	host.Host
}

// Title - self-explanatory.
func (l ListItemHost) Title() string { return l.Host.Title }

// Description - self-explanatory.
func (l ListItemHost) Description() string { return l.Host.Description }

// FilterValue - returns the field combination which are used when user performs a search in the list.
func (l ListItemHost) FilterValue() string { return l.Host.Title + l.Host.Description }

// CompareTo - compares this listItemHost with another one.
func (l ListItemHost) CompareTo(host ListItemHost) int {
	if l.Host.Title == host.Title() {
		return l.Host.ID - host.ID
	}

	return lo.Ternary(l.Host.Title < host.Title(), -1, 1)
}
