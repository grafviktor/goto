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
	"github.com/grafviktor/goto/internal/utils"
)

var groupHint = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#9B9B9B", Dark: "#5C5C5C"})

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
			switch *delegate.layout {
			case constant.ScreenLayoutTight:
				*delegate.layout = constant.ScreenLayoutNormal
			case constant.ScreenLayoutNormal:
				*delegate.layout = constant.ScreenLayoutGroup
			default:
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

func (hd *hostDelegate) isHostMovedToAnotherGroup(hostGroup string) bool {
	appStatedGroup := hd.selectedGroup
	return appStatedGroup != nil &&
		*appStatedGroup != "" &&
		!strings.EqualFold(hostGroup, *appStatedGroup)
}

func (hd *hostDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	hostItem, ok := item.(ListItemHost)
	if *hd.layout == constant.ScreenLayoutGroup {
		hostItem.Host.Description = hostItem.Group
	} else if ok && hd.isHostMovedToAnotherGroup(hostItem.Group) {
		groupIsEmpty := utils.StringEmpty(hostItem.Group)
		groupName := lo.Ternary(groupIsEmpty, "[no group]", fmt.Sprintf("(%s)", hostItem.Group))
		hostItem.Host.Title = fmt.Sprintf("%s %s", hostItem.Title(), groupHint.Render(groupName))
	}

	hd.DefaultDelegate.Render(w, m, index, hostItem)
}
