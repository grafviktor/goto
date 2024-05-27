// Package constant contains shared app constants
package constant

import "errors"

// ErrNotFound is used by data layer.
var ErrNotFound = errors.New("not found")

// ProtocolSSH - is only supported protocol.
const ProtocolSSH = "ssh"

// ScreenLayout is used to determine how the hostlist should be displayed.
type ScreenLayout string

const (
	// LayoutTight is set when all items in the hostlist are shown without description.
	LayoutTight ScreenLayout = "tight"
	// LayoutNormal is set when all hosts are shown with description field and a margin.
	LayoutNormal ScreenLayout = "normal"
)
