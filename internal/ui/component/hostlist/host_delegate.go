package hostlist

import (
	"io"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/grafviktor/goto/internal/constant"
)

type hostDelegate struct {
	list.DefaultDelegate
	layout *constant.ScreenLayout
	logger iLogger
}

// NewHostDelegate creates a new Delegate object which can be used for customizing the view of a host.
func NewHostDelegate(layout *constant.ScreenLayout, log iLogger) *hostDelegate {
	delegate := &hostDelegate{
		DefaultDelegate: list.NewDefaultDelegate(),
		logger:          log,
		layout:          layout,
	}

	delegate.updateLayout()

	delegate.UpdateFunc = func(msg tea.Msg, m *list.Model) tea.Cmd {
		if _, ok := msg.(msgToggleLayout); ok {
			if *delegate.layout == constant.ScreenLayoutTight {
				*delegate.layout = constant.ScreenLayoutNormal
			} else {
				// If layout is not set or "Normal", switch to "tight" layout.
				*delegate.layout = constant.ScreenLayoutTight
			}

			delegate.updateLayout()
		}

		return nil
	}

	return delegate
}

func (hd *hostDelegate) updateLayout() {
	if *hd.layout == constant.ScreenLayoutTight {
		hd.SetSpacing(0)
		hd.ShowDescription = false
	} else {
		// If layout is not set or "Normal", switch to "tight" layout.
		hd.SetSpacing(1)
		hd.ShowDescription = true
	}

	hd.logger.Debug("[UI] Change screen layout to: '%s'", *hd.layout)
}

func (hd *hostDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	var hostItem ListItemHost
	// FIXME:
	// Should greyout items from a different group.
	// Should check selected group and compare with the item's group instead of the layout
	if (hd.layout != nil) {
		var ok bool
		if hostItem, ok = item.(ListItemHost); ok {
			greyedOutStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#999999"));

			hostItem.Host.Title = greyedOutStyle.Render(hostItem.Title());
		}
	}

	hd.DefaultDelegate.Render(w, m, index, hostItem)
}