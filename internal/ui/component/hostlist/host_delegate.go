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

func (hd *hostDelegate) updateLayout() {
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

func (hd *hostDelegate) isHostMovedToAnotherGroup(hostGroup string) bool {
	// Return false if group is not selected, as all hosts should be displayed.
	if utils.StringEmpty(hd.selectedGroup) {
		return false
	}

	return !strings.EqualFold(*hd.selectedGroup, hostGroup)
}

func (hd *hostDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	if itemCopy, ok := item.(ListItemHost); ok {
		if hd.layout != nil && *hd.layout == constant.ScreenLayoutGroup {
			itemCopy.Host.Description = itemCopy.Group
		} else if hd.isHostMovedToAnotherGroup(itemCopy.Group) {
			groupIsEmpty := utils.StringEmpty(&itemCopy.Group)
			groupName := lo.Ternary(groupIsEmpty, "[no group]", fmt.Sprintf("(%s)", itemCopy.Group))
			itemCopy.Host.Title = fmt.Sprintf("%s %s", itemCopy.Title(), groupHint.Render(groupName))
		}

		hd.DefaultDelegate.Render(w, m, index, itemCopy)
	} else {
		hd.DefaultDelegate.Render(w, m, index, item)
	}
}
