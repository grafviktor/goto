package hostlist

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/samber/lo"

	"github.com/grafviktor/goto/internal/constant"
	"github.com/grafviktor/goto/internal/utils"
)

type HostDelegate struct {
	list.DefaultDelegate

	layout        *constant.ScreenLayout
	selectedGroup *string
	logger        iLogger
	styles        styles
}

// NewHostDelegate creates a new Delegate object which can be used for customizing the view of a host.
func NewHostDelegate(layout *constant.ScreenLayout, group *string, log iLogger) *HostDelegate {
	delegate := &HostDelegate{
		DefaultDelegate: list.NewDefaultDelegate(),
		logger:          log,
		layout:          layout,
		selectedGroup:   group,
		styles:          defaultStyles(),
	}

	delegate.Styles = delegate.styles.listDelegate
	delegate.updateLayout()

	delegate.UpdateFunc = func(msg tea.Msg, _ *list.Model) tea.Cmd {
		if _, ok := msg.(msgToggleLayout); ok {
			//exhaustive:ignore
			switch *delegate.layout {
			case constant.ScreenLayoutCompact:
				*delegate.layout = constant.ScreenLayoutDescription
			case constant.ScreenLayoutDescription:
				*delegate.layout = constant.ScreenLayoutGroup
			default:
				*delegate.layout = constant.ScreenLayoutCompact
			}

			delegate.updateLayout()
		}

		return nil
	}

	return delegate
}

func (hd *HostDelegate) updateLayout() {
	if *hd.layout == constant.ScreenLayoutCompact {
		hd.SetSpacing(0)
		hd.ShowDescription = false
	} else {
		// If layout is not set or "Normal", switch to "tight" layout.
		hd.SetSpacing(1)
		hd.ShowDescription = true
	}

	hd.logger.Debug("[UI] Change screen layout to: '%s'", *hd.layout)
}

func (hd *HostDelegate) isHostMovedToAnotherGroup(hostGroup string) bool {
	// Return false if group is not selected, as all hosts should be displayed.
	if utils.StringEmpty(hd.selectedGroup) {
		return false
	}

	return !strings.EqualFold(*hd.selectedGroup, hostGroup)
}

func (hd *HostDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	if itemCopy, ok := item.(ListItemHost); ok {
		if hd.layout != nil && *hd.layout == constant.ScreenLayoutGroup {
			itemCopy.Host.Description = itemCopy.Group
		} else if hd.isHostMovedToAnotherGroup(itemCopy.Group) {
			groupIsEmpty := utils.StringEmpty(&itemCopy.Group)
			groupName := lo.Ternary(groupIsEmpty, "[no group]", fmt.Sprintf("(%s)", itemCopy.Group))
			itemCopy.Host.Title = fmt.Sprintf("%s %s", itemCopy.Title(), hd.styles.groupHint.Render(groupName))
		}

		hd.DefaultDelegate.Render(w, m, index, itemCopy)
	} else {
		hd.DefaultDelegate.Render(w, m, index, item)
	}
}
