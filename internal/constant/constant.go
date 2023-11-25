// Package constant contains shared app constants
package constant

import "errors"

// ErrNotFound is used by data layer.
var ErrNotFound = errors.New("not found")

// ErrDuplicateRecord = errors.New("duplicate record")
// ErrDeleted         = errors.New("deleted")
// ErrNoUserID        = errors.New("no user ID")
// ErrBadArgument     = errors.New("bad argument")

// ProtocolSSH - is only supported protocol.
const ProtocolSSH = "ssh"
