package splitview

import (
	"charm.land/lipgloss/v2"
	"github.com/grafviktor/goto/internal/ui/theme"
)

type styles struct {
	divider lipgloss.Style
}

func defaultStyles() styles {
	theme := theme.Get().Styles.SplitView

	return styles{
		divider: theme.Divider,
	}
}
