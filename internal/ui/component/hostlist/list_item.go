package hostlist

import "github.com/grafviktor/goto/internal/model"

// ListItemHost is an adaptor between host model and bubbletea list model.
type ListItemHost struct {
	model.Host
}

// Title - self-explanatory.
func (l ListItemHost) Title() string { return l.Host.Title }

// Description - self-explanatory.
func (l ListItemHost) Description() string { return l.Host.Description }

// FilterValue - returns the field combination which are used when user performs a search in the list.
func (l ListItemHost) FilterValue() string { return l.Host.Title + l.Host.Description }

// Unwrap - extracts host model from list item.
func (l ListItemHost) Unwrap() *model.Host { return &l.Host }
