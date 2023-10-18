package hostlist

import "github.com/grafviktor/goto/internal/model"

type ListItemHost struct {
	model.Host
}

func (l ListItemHost) Title() string       { return l.Host.Title }
func (l ListItemHost) Description() string { return l.Host.Description }
func (l ListItemHost) FilterValue() string { return l.Host.Title + l.Host.Description }
func (l ListItemHost) Unwrap() *model.Host { return &l.Host }
