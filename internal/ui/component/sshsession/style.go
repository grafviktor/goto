package sshsession

import (
	"charm.land/lipgloss/v2"

	"github.com/grafviktor/goto/internal/ui/theme"
)

type styles struct {
	statusLine lipgloss.Style
}

func defaultStyles() styles {
	theme := theme.Get().Styles.SSHSession

	return styles{
		statusLine: theme.StatusLine,
	}
}
