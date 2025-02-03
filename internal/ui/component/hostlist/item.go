package hostlist

import (
	"fmt"

	"github.com/grafviktor/goto/internal/model/host"
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

// uniqueName - generates a unique name for a host.
// Unique name is used to identify a host position in the list.
// It's a naive implementation, but good enough for the current use case.
//   - host:
//     title: SomeHost
//     id: 5
//     # => SomeHost00005
func (l ListItemHost) uniqueName() string {
	return fmt.Sprintf("%s%05d", l.Host.Title, l.Host.ID)
}
