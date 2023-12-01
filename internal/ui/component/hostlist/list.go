// Package hostlist implements the host list view.
package hostlist

import (
	"context"
	"errors"
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/exp/slices"

	"github.com/grafviktor/goto/internal/model"
	"github.com/grafviktor/goto/internal/state"
	"github.com/grafviktor/goto/internal/storage"
	"github.com/grafviktor/goto/internal/ui/message"
	"github.com/grafviktor/goto/internal/utils"
	"github.com/grafviktor/goto/internal/utils/ssh"
)

var (
	docStyle               = lipgloss.NewStyle().Margin(1, 2)
	itemNotSelectedMessage = "you must select an item"
)

type logger interface {
	Debug(format string, args ...any)
}

type (
	// MsgEditItem fires when user press edit button.
	MsgEditItem struct{ HostID int }
	// MsgCopyItem fires when user press copy button.
	MsgCopyItem struct{ HostID int }
	// MsgSelectItem is required to let host list know that it's time to update title.
	MsgSelectItem struct{ HostID int }
	// MsgNewItem fires when user press new host button.
	MsgNewItem      struct{}
	msgInitComplete struct{}
	msgErrorOccured struct{ err error }
	// MsgRepoUpdated - fires when data layer updated and it's required to reload the host list.
	MsgRepoUpdated  struct{}
	msgFocusChanged struct{}
)

type listModel struct {
	innerModel list.Model
	repo       storage.HostStorage
	keyMap     *keyMap
	appState   *state.ApplicationState
	logger     logger
}

// New - creates new host list model.
// context - is not used.
// storage - is the data layer.
// appState - is the application state, usually we want to restore previous state when application restarts,
// for instance focus previously selected host.
// log - application logger.
func New(_ context.Context, storage storage.HostStorage, appState *state.ApplicationState, log logger) listModel {
	delegate := list.NewDefaultDelegate()
	delegateKeys := newDelegateKeyMap()
	listItems := []list.Item{}
	m := listModel{
		innerModel: list.New(listItems, delegate, 0, 0),
		keyMap:     delegateKeys,
		repo:       storage,
		appState:   appState,
		logger:     log,
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

func (m listModel) Init() tea.Cmd {
	return tea.Batch(message.TeaCmd(msgInitComplete{}))
}

func (m listModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// dispatch msgFocusChanged message to update list title
		cmds = append(cmds, message.TeaCmd(msgFocusChanged{}))

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
			return m, message.TeaCmd(MsgEditItem{})
		case key.Matches(msg, m.keyMap.clone):
			return m.copyItem(msg)
		}
	case tea.WindowSizeMsg:
		// triggers immediately after app start because we render this component by default
		h, v := docStyle.GetFrameSize()
		m.innerModel.SetSize(msg.Width-h, msg.Height-v)
		m.logger.Debug("Set host list size: %d %d", m.innerModel.Width(), m.innerModel.Height())
	case msgErrorOccured:
		return m.listTitleUpdate(msg), nil
	case MsgRepoUpdated:
		return m.refreshRepo(msg)
	case msgInitComplete:
		return m.refreshRepo(msg)
	case msgFocusChanged:
		m = m.listTitleUpdate(msg)
		var cmd tea.Cmd
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

func (m listModel) View() string {
	return docStyle.Render(m.innerModel.View())
}

func (m listModel) removeItem(_ tea.Msg) (listModel, tea.Cmd) {
	item, ok := m.innerModel.SelectedItem().(ListItemHost)
	if !ok {
		return m, message.TeaCmd(msgErrorOccured{err: errors.New("you must select an item")})
	}

	err := m.repo.Delete(item.ID)
	if err != nil {
		return m, message.TeaCmd(msgErrorOccured{err})
	}

	return m, tea.Batch(
		message.TeaCmd(MsgRepoUpdated{}),
		message.TeaCmd(msgFocusChanged{}),
	)
}

func (m listModel) refreshRepo(_ tea.Msg) (listModel, tea.Cmd) {
	items := []list.Item{}
	hosts, err := m.repo.GetAll()
	if err != nil {
		return m, message.TeaCmd(msgErrorOccured{err})
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

	return m, tea.Batch(setItemsCmd, message.TeaCmd(msgFocusChanged{}))
}

func (m listModel) editItem(_ tea.Msg) (listModel, tea.Cmd) {
	item, ok := m.innerModel.SelectedItem().(ListItemHost)
	if !ok {
		return m, message.TeaCmd(msgErrorOccured{err: errors.New(itemNotSelectedMessage)})
	}

	host := *item.Unwrap()
	return m, message.TeaCmd(MsgEditItem{HostID: host.ID})
}

func (m listModel) copyItem(_ tea.Msg) (listModel, tea.Cmd) {
	item, ok := m.innerModel.SelectedItem().(ListItemHost)
	if !ok {
		return m, message.TeaCmd(msgErrorOccured{err: errors.New(itemNotSelectedMessage)})
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
		return m, message.TeaCmd(msgErrorOccured{err})
	}

	return m, tea.Batch(
		message.TeaCmd(MsgRepoUpdated{}),
		message.TeaCmd(msgFocusChanged{}),
	)
}

func (m listModel) executeCmd(_ tea.Msg) (listModel, tea.Cmd) {
	item, ok := m.innerModel.SelectedItem().(ListItemHost)
	if !ok {
		return m, message.TeaCmd(msgErrorOccured{err: errors.New(itemNotSelectedMessage)})
	}

	host := *item.Unwrap()
	err := m.repo.Save(host)
	if err != nil {
		return m, message.TeaCmd(msgErrorOccured{err})
	}

	command := ssh.ConstructCMD(ssh.BaseCMD(), utils.HostModelToOptionsAdaptor(host)...)
	process := utils.BuildProcess(command)
	return m, tea.ExecProcess(process, func(err error) tea.Msg {
		if err != nil {
			return msgErrorOccured{err}
		}

		return nil
	})
}

func (m listModel) listTitleUpdate(msg tea.Msg) listModel {
	switch msg := msg.(type) {
	case msgErrorOccured:
		m.innerModel.Title = msg.err.Error()

		return m
	default:
		item, ok := m.innerModel.SelectedItem().(ListItemHost)
		if !ok {
			return m
		}

		m.innerModel.Title = ssh.ConstructCMD("ssh", utils.HostModelToOptionsAdaptor(*item.Unwrap())...)

		return m
	}
}

func (m listModel) onFocusChanged(_ tea.Msg) (listModel, tea.Cmd) {
	if hostItem, ok := m.innerModel.SelectedItem().(ListItemHost); ok {
		return m, message.TeaCmd(MsgSelectItem{HostID: hostItem.ID})
	}

	return m, nil
}
