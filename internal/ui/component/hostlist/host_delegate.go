package hostlist

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"

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
			if *delegate.layout == constant.LayoutTight {
				*delegate.layout = constant.LayoutNormal
			} else {
				// If layout is not set or "Normal", switch to "tight" layout.
				*delegate.layout = constant.LayoutTight
			}

			delegate.updateLayout()
		}

		return nil
	}

	return delegate
}

func (hd *hostDelegate) updateLayout() {
	if *hd.layout == constant.LayoutTight {
		hd.SetSpacing(0)
		hd.ShowDescription = false
	} else {
		// If layout is not set or "Normal", switch to "tight" layout.
		hd.SetSpacing(1)
		hd.ShowDescription = true
	}

	hd.logger.Debug("[UI] Change screen layout to: '%s'", *hd.layout)
}
