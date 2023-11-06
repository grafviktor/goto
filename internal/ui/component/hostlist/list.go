package hostlist

import (
	"context"
	"errors"
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/grafviktor/goto/internal/connector/ssh"
	"github.com/grafviktor/goto/internal/model"
	"github.com/grafviktor/goto/internal/state"
	"github.com/grafviktor/goto/internal/storage"
	. "github.com/grafviktor/goto/internal/ui/message" //nolint dot-imports
	"golang.org/x/exp/slices"
)

var docStyle = lipgloss.NewStyle().Margin(1, 2)

type logger interface {
	Debug(format string, args ...any)
}

type (
	MsgEditItem     struct{ HostID int }
	MsgCopyItem     struct{ HostID int }
	MsgSelectItem   struct{ HostID int }
	MsgNewItem      struct{}
	msgInitComplete struct{}
	msgErrorOccured struct{ err error }
	MsgRepoUpdated  struct{}
	msgFocusChanged struct{}
)

type ListModel struct {
	innerModel list.Model
	repo       storage.HostStorage
	keyMap     *keyMap
	appState   *state.ApplicationState
	logger     logger
	Created    string
}

func New(_ context.Context, storage storage.HostStorage, appState *state.ApplicationState, log logger) ListModel {
	delegate := list.NewDefaultDelegate()
	delegateKeys := newDelegateKeyMap()
	listItems := []list.Item{}
	m := ListModel{
		innerModel: list.New(listItems, delegate, 0, 0),
		keyMap:     delegateKeys,
		repo:       storage,
		appState:   appState,
		logger:     log,
		Created:    "A",
	}

	m.innerModel.KeyMap.CursorUp.Unbind()
	m.innerModel.KeyMap.CursorUp = delegateKeys.cursorUp
	m.innerModel.KeyMap.CursorDown.Unbind()
	m.innerModel.KeyMap.CursorDown = delegateKeys.cursorDown

	m.innerModel.AdditionalShortHelpKeys = delegateKeys.ShortHelp
	m.innerModel.AdditionalFullHelpKeys = delegateKeys.FullHelp

	m.innerModel.Title = "goto:"
	m.innerModel.SetShowStatusBar(false)

	return m
}

func (m ListModel) Init() tea.Cmd {
	return tea.Batch(TeaCmd(msgInitComplete{}))
}

func (m ListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// dispatch msgFocusChanged message to update list title
		cmds = append(cmds, TeaCmd(msgFocusChanged{}))

		if m.innerModel.FilterState() == list.Filtering {
			// if filter is enabled, we should not handle any keyboard messages
			break
		}

		switch {
		case key.Matches(msg, m.keyMap.connect):
			return m.executeCmd(msg)
		case key.Matches(msg, m.keyMap.remove):
			return m.removeItem(msg)
		case key.Matches(msg, m.keyMap.edit):
			return m.editItem(msg)
		case key.Matches(msg, m.keyMap.append):
			return m, TeaCmd(MsgEditItem{})
		case key.Matches(msg, m.keyMap.clone):
			return m.copyItem(msg)
		}
	case tea.WindowSizeMsg:
		// triggers immediately after app start because we render this component by default
		h, v := docStyle.GetFrameSize()
		m.innerModel.SetSize(msg.Width-h, msg.Height-v)
		m.logger.Debug("New frame size: %d %d", m.innerModel.Width(), m.innerModel.Height())
	case msgErrorOccured:
		return m.listTitleUpdate(msg)
	case MsgRepoUpdated:
		return m.refreshRepo(msg)
	case msgInitComplete:
		return m.refreshRepo(msg)
	case msgFocusChanged:
		var cmd tea.Cmd
		m, cmd = m.listTitleUpdate(msg)
		cmds = append(cmds, cmd)
		m, cmd = m.onFocusChanged(msg)
		cmds = append(cmds, cmd)
		return m, tea.Batch(cmds...)
	}

	// If we could not find our own update handler, we pass message to the original model
	// otherwise we would have to implement all key hanlders and other stuff by ourselves
	var innerModelCmd tea.Cmd
	m.innerModel, innerModelCmd = m.innerModel.Update(msg)
	cmds = append(cmds, innerModelCmd)
	return m, tea.Batch(cmds...)
}

func (m ListModel) View() string {
	return docStyle.Render(m.innerModel.View())
}

func (m ListModel) removeItem(_ tea.Msg) (ListModel, tea.Cmd) {
	item, ok := m.innerModel.SelectedItem().(ListItemHost)
	if !ok {
		return m, TeaCmd(msgErrorOccured{err: errors.New("You must select an item")})
	}
	m.repo.Delete(item.ID)

	return m, tea.Batch(TeaCmd(MsgRepoUpdated{}), TeaCmd(msgFocusChanged{}))
}

func (m ListModel) refreshRepo(_ tea.Msg) (ListModel, tea.Cmd) {
	items := []list.Item{}
	hosts, err := m.repo.GetAll()
	if err != nil {
		return m, TeaCmd(msgErrorOccured{err})
	}

	slices.SortFunc(hosts, func(a, b model.Host) int {
		if a.Title < b.Title {
			return -1
		}
		return 1
	})

	// Wrap hosts into List items
	for _, h := range hosts {
		items = append(items, ListItemHost{Host: h})
	}

	setItemsCmd := m.innerModel.SetItems(items)

	// we restore selected item from application configuration
	for uiIndex, listItem := range m.innerModel.VisibleItems() {
		if hostItem, ok := listItem.(ListItemHost); ok {
			if m.appState.Selected == hostItem.ID {
				m.innerModel.Select(uiIndex)
				break
			}
		}
	}

	return m, tea.Batch(setItemsCmd, TeaCmd(msgFocusChanged{}))
}

func (m ListModel) editItem(_ tea.Msg) (ListModel, tea.Cmd) {
	item, ok := m.innerModel.SelectedItem().(ListItemHost)
	if !ok {
		errText := "You must select an item"
		return m, TeaCmd(msgErrorOccured{err: errors.New(errText)})
	}

	host := *item.Unwrap()
	return m, TeaCmd(MsgEditItem{HostID: host.ID})
}

func (m ListModel) copyItem(_ tea.Msg) (ListModel, tea.Cmd) {
	item, ok := m.innerModel.SelectedItem().(ListItemHost)
	if !ok {
		errText := "You must select an item"
		return m, TeaCmd(msgErrorOccured{err: errors.New(errText)})
	}

	originalHost := item.Unwrap()
	clonedHost := originalHost.Clone()
	for i := 1; ok; i++ {
		clonedHostTitle := fmt.Sprintf("%s %d", originalHost.Title, i)
		listItems := m.innerModel.Items()
		idx := slices.IndexFunc(listItems, func(li list.Item) bool {
			return li.(ListItemHost).Title() == clonedHostTitle
		})

		if idx < 0 {
			clonedHost.Title = clonedHostTitle
			break
		}
	}

	if err := m.repo.Save(clonedHost); err != nil {
		return m, TeaCmd(msgErrorOccured{err})
	}

	return m, tea.Batch(TeaCmd(MsgRepoUpdated{}), TeaCmd(msgFocusChanged{}))
}

func (m ListModel) executeCmd(_ tea.Msg) (ListModel, tea.Cmd) {
	item, ok := m.innerModel.SelectedItem().(ListItemHost)
	if !ok {
		errText := "You must select an item"

		return m, TeaCmd(msgErrorOccured{err: errors.New(errText)})
	}

	host := *item.Unwrap()
	err := m.repo.Save(host)
	if err != nil {
		return m, TeaCmd(msgErrorOccured{err})
	}

	connectSSHCmd := ssh.Connect(host)
	return m, tea.ExecProcess(connectSSHCmd, func(err error) tea.Msg {
		if err != nil {
			return msgErrorOccured{err}
		}

		return nil
	})
}

func (m ListModel) listTitleUpdate(msg tea.Msg) (ListModel, tea.Cmd) {
	switch msg := msg.(type) {
	case msgErrorOccured:
		m.innerModel.Title = fmt.Sprintf("%s", msg.err.Error())

		return m, nil
	default:
		item, ok := m.innerModel.SelectedItem().(ListItemHost)
		if !ok {
			return m, nil
		}

		m.innerModel.Title = fmt.Sprintf("goto: %s", item.Unwrap().Address)

		return m, nil
	}
}

func (m ListModel) onFocusChanged(msg tea.Msg) (ListModel, tea.Cmd) {
	if hostItem, ok := m.innerModel.SelectedItem().(ListItemHost); ok {
		return m, TeaCmd(MsgSelectItem{HostID: hostItem.ID})
	}

	return m, nil
}
