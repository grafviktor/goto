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

type (
	MsgEditItem     struct{ HostID int }
	MsgCopyItem     struct{ HostID int }
	MsgSelectItem   struct{ ID int }
	MsgNewItem      struct{}
	msgInitComplete struct{}
	msgErrorOccured struct{ err error }
	MsgRepoUpdated  struct{}
	msgFocusChanged struct{}
)

type listModel struct {
	innerModel list.Model
	repo       storage.HostStorage
	keyMap     *keyMap
	appState   *state.ApplicationState
}

func New(_ context.Context, storage storage.HostStorage, appState *state.ApplicationState) listModel {
	delegate := list.NewDefaultDelegate()
	delegateKeys := newDelegateKeyMap()
	listItems := []list.Item{}
	m := listModel{
		innerModel: list.New(listItems, delegate, 0, 0),
		keyMap:     delegateKeys,
		repo:       storage,
		appState:   appState,
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
	return tea.Batch(TeaCmd(msgInitComplete{}))
}

func (m listModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
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

		// dispatch msgFocusChanged message to update list title
		cmds = append(cmds, TeaCmd(msgFocusChanged{}))
	case tea.WindowSizeMsg:
		// triggers immediately after app start because we render this component by default
		h, v := docStyle.GetFrameSize()
		m.innerModel.SetSize(msg.Width-h, msg.Height-v)
	case msgErrorOccured:
		return m.listTitleUpdate(msg)
	case MsgRepoUpdated:
		return m.refreshRepo(msg)
	case msgInitComplete:
		return m.refreshRepo(msg)
	case msgFocusChanged:
		m, cmd := m.listTitleUpdate(msg)
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

func (m listModel) View() string {
	return docStyle.Render(m.innerModel.View())
}

func (m listModel) removeItem(_ tea.Msg) (listModel, tea.Cmd) {
	item, ok := m.innerModel.SelectedItem().(ListItemHost)
	if !ok {
		return m, TeaCmd(msgErrorOccured{err: errors.New("You must select an item")})
	}
	m.repo.Delete(item.ID)

	return m, tea.Batch(TeaCmd(MsgRepoUpdated{}), TeaCmd(msgFocusChanged{}))
}

func (m listModel) refreshRepo(_ tea.Msg) (listModel, tea.Cmd) {
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

	m.innerModel.SetItems(items)

	// we restore selected item from application configuration
	for uiIndex, listItem := range m.innerModel.Items() {
		if hostItem, ok := listItem.(ListItemHost); ok {
			if m.appState.Selected == hostItem.ID {
				m.innerModel.Select(uiIndex)
				break
			}
		}
	}

	return m, TeaCmd(msgFocusChanged{})
}

func (m listModel) editItem(_ tea.Msg) (listModel, tea.Cmd) {
	item, ok := m.innerModel.SelectedItem().(ListItemHost)
	if !ok {
		errText := "You must select an item"
		return m, TeaCmd(msgErrorOccured{err: errors.New(errText)})
	}

	host := *item.Unwrap()
	return m, TeaCmd(MsgEditItem{HostID: host.ID})
}

func (m listModel) copyItem(_ tea.Msg) (listModel, tea.Cmd) {
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

func (m listModel) executeCmd(_ tea.Msg) (listModel, tea.Cmd) {
	item, ok := m.innerModel.SelectedItem().(ListItemHost)
	if !ok {
		errText := "could not cast list.SelectedItem() to component.ListItem"

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

func (m listModel) listTitleUpdate(msg tea.Msg) (listModel, tea.Cmd) {
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

func (m listModel) onFocusChanged(msg tea.Msg) (listModel, tea.Cmd) {
	if hostItem, ok := m.innerModel.SelectedItem().(ListItemHost); ok {
		return m, TeaCmd(MsgSelectItem{ID: hostItem.ID})
	}

	return m, nil
}
