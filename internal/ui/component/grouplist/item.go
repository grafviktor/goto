package grouplist

// ListItemHostGroup is an adaptor between group model and bubbletea list model.
type ListItemHostGroup struct {
	groupName string
}

// Title - self-explanatory.
func (l ListItemHostGroup) Title() string { return l.groupName }

// Description - self-explanatory.
func (l ListItemHostGroup) Description() string { return l.groupName }

// FilterValue - returns the field combination which are used when user performs a search in the list.
func (l ListItemHostGroup) FilterValue() string { return l.groupName }