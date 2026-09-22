package sshsession

import (
	"charm.land/lipgloss/v2"

	"github.com/grafviktor/goto/internal/ui/theme"
)

type styles struct {
	header     lipgloss.Style
	statusText lipgloss.Style
}

func defaultStyles() styles {
	themeSettings := theme.Get().Styles.SSHSession

	return styles{
		header:     themeSettings.Header,
		statusText: themeSettings.StatusText,
	}
}
