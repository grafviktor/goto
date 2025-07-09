// Package test contains utility methods for unit tests.
package testutils_test

import (
	"reflect"

	tea "github.com/charmbracelet/bubbletea"
)

// CmdToMessage - should only be used in unit tests.
func CmdToMessage(cmd tea.Cmd, messages *[]tea.Msg) {
	if cmd == nil {
		// Probably, there is "nil" instead of a real message. Returning.
		return
	}

	message := cmd()
	valueOf := reflect.ValueOf(message)

	// Slice of messages is returned by tea.BatchMsg or tea.sequenceMsg
	// See also github.com/charmbracelet/bubbletea@v0.25.0/tea.go#eventLoop
	// which handles tea.Sequence
	if valueOf.Kind() == reflect.Slice {
		for i := 0; i < valueOf.Len(); i++ {
			if valueOf.Index(i).Kind() == reflect.Func {
				// If it's a function, then it's probably inner tea.Cmd
				innerCmd, _ := valueOf.Index(i).Interface().(tea.Cmd)
				CmdToMessage(innerCmd, messages)
			} else {
				// Otherwise it's a slice of real messages
				*messages = append(*messages, message)
			}
		}
	} else {
		*messages = append(*messages, message)
	}
}
