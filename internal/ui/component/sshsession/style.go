package sshsession

import (
	"charm.land/lipgloss/v2"

	"github.com/grafviktor/goto/internal/ui/theme"
)

type styles struct {
	groupColor lipgloss.Style
	hostColor  lipgloss.Style
}

func defaultStyles() styles {
	themeSettings := theme.Get().Styles.SSHSession

	return styles{
		groupColor: themeSettings.GroupColor,
		hostColor:  themeSettings.HostColor,
	}
}
