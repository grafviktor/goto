package constant

import "errors"

var (
	ErrNotFound    = errors.New("not found")
	ErrBadArgument = errors.New("bad argument")
)

// ErrDuplicateRecord = errors.New("duplicate record")
// ErrDeleted         = errors.New("deleted")
// ErrNoUserID        = errors.New("no user ID")
// ErrBadArgument     = errors.New("bad argument")

const ProtocolSSH = "ssh"
