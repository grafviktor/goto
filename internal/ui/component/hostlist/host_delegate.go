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

// var selectedStyle = lipgloss.NewStyle().
// 	Border(lipgloss.NormalBorder(), false, true, false, false).
// 	BorderForeground(lipgloss.AdaptiveColor{Light: "#F793FF", Dark: "#AD58B4"}).
// 	Foreground(lipgloss.AdaptiveColor{Light: "#EE6FF8", Dark: "#EE6FF8"}).
// 	Padding(0, 0, 0, 0)

func (hd *hostDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	// TODO:
	// 1. Make an abbreviation from group name My Group -> MG
	// 2. If group is not selected, display group abbreviation in grey color before title
	// Also see https://github.com/charmbracelet/bubbletea/blob/main/examples/list-simple/main.go

	defaultRenderer := hd.DefaultDelegate.Render
	hostItem, ok := item.(ListItemHost)
	if ok {
		// TODO: Refactor!
		if hd.selectedGroup != nil &&
			*hd.selectedGroup != "" &&
			!strings.EqualFold(hostItem.Group, *hd.selectedGroup) {
			greyedOutStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#333333"))
			groupName := lo.Ternary(strings.TrimSpace(hostItem.Group) == "",
				"[group removed]",
				fmt.Sprintf("['%s']", hostItem.Group))
			// groupMessage := fmt.Sprintf("['%s']", hostItem.Group)
			title := fmt.Sprintf("%s %s", hostItem.Title(), greyedOutStyle.Render(groupName))
			hostItem.Host.Title = title
		}
	}

	// Use buffered writer to hack the default renderer output
	// bw := utils.ProcessBufferWriter{}
	// defaultRenderer(&bw, m, index, hostItem)
	// text := string(bw.Output)
	// fmt.Fprint(w, cutItemBorder(text))

	defaultRenderer(w, m, index, hostItem)
}

/* func cutItemBorder(str string) string {
	// runes := []rune(str)
	// if len(runes) < 2 {
	// 	return str
	// }
	resetStyle := lipgloss.NewStyle()

	if len(str) < 3 {
		return str
	}

	return resetStyle.Render(str)
} */
