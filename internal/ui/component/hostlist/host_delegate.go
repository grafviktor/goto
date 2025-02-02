package hostlist

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/samber/lo"

	"github.com/grafviktor/goto/internal/constant"
)

type hostDelegate struct {
	list.DefaultDelegate
	layout        *constant.ScreenLayout
	selectedGroup *string
	logger        iLogger
}

// NewHostDelegate creates a new Delegate object which can be used for customizing the view of a host.
func NewHostDelegate(layout *constant.ScreenLayout, group *string, log iLogger) *hostDelegate {
	delegate := &hostDelegate{
		DefaultDelegate: list.NewDefaultDelegate(),
		logger:          log,
		layout:          layout,
		selectedGroup:   group,
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
	if hd.layout != nil {
		var ok bool
		if hostItem, ok = item.(ListItemHost); ok {
			// TODO: Refactor!
			if hd.selectedGroup != nil &&
				*hd.selectedGroup != "" &&
				!strings.EqualFold(hostItem.Group, *hd.selectedGroup) {
				greyedOutStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#333"))
				groupName := lo.Ternary(strings.TrimSpace(hostItem.Group) == "",
					"[n/a]",
					fmt.Sprintf("['%s']", hostItem.Group))
				// groupMessage := fmt.Sprintf("['%s']", hostItem.Group)
				title := fmt.Sprintf("%s %s", hostItem.Title(), greyedOutStyle.Render(groupName))
				hostItem.Host.Title = title
			}
		}
	}

	hd.DefaultDelegate.Render(w, m, index, hostItem)
}
