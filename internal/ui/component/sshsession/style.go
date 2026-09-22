package sshsession

import (
	"charm.land/lipgloss/v2"

	"github.com/grafviktor/goto/internal/ui/theme"
)

type styles struct {
	header lipgloss.Style
}

func defaultStyles() styles {
	theme := theme.Get().Styles.SSHSession

	return styles{
		header: theme.Header,
	}
}
