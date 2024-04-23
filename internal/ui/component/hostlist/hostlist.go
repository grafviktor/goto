// Package hostlist implements the host list view.
package hostlist

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/exp/slices"

	"github.com/grafviktor/goto/internal/model"
	"github.com/grafviktor/goto/internal/state"
	"github.com/grafviktor/goto/internal/storage"
	"github.com/grafviktor/goto/internal/ui/message"
)

var (
	docStyle               = lipgloss.NewStyle().Margin(1, 2)
	itemNotSelectedMessage = "you must select an item"
	modeRemoveItem         = "removeItem"
	modeDefault            = ""
	defaultListTitle       = "press 'n' to add a new host"
)

type iLogger interface {
	Debug(format string, args ...any)
	Info(format string, args ...any)
	Error(format string, args ...any)
}

type (
	// OpenEditForm fires when user press edit button.
	OpenEditForm struct{ HostID int }
	// MsgCopyItem fires when user press copy button.
	MsgCopyItem      struct{ HostID int }
	msgErrorOccurred struct{ err error }
	// MsgRefreshRepo - fires when data layer updated, and it's required to reload the host list.
	MsgRefreshRepo struct{}
	msgRefreshUI   struct{}
)

type listModel struct {
	innerModel list.Model
	repo       storage.HostStorage
	keyMap     *keyMap
	appState   *state.ApplicationState
	logger     iLogger
	mode       string
}

// New - creates new host list model.
// context - is not used.
// storage - is the data layer.
// appState - is the application state, usually we want to restore previous state when application restarts,
// for instance focus previously selected host.
// log - application logger.
func New(_ context.Context, storage storage.HostStorage, appState *state.ApplicationState, log iLogger) *listModel {
	delegate := list.NewDefaultDelegate()
	delegateKeys := newDelegateKeyMap()

	var listItems []list.Item
	innerModel := list.New(listItems, delegate, 0, 0)
	// This line affects sorting when filtering enabled. What UnsortedFilter
	// does - it filters the collection, but leaves initial items order unchanged.
	// Default filter on the contrary - filters the collection based on the match rank.
	innerModel.Filter = list.UnsortedFilter

	m := listModel{
		innerModel: innerModel,
		keyMap:     delegateKeys,
		repo:       storage,
		appState:   appState,
		logger:     log,
	}

	m.innerModel.KeyMap.CursorUp.Unbind()
	m.innerModel.KeyMap.CursorUp = delegateKeys.cursorUp
	m.innerModel.KeyMap.CursorDown.Unbind()
	m.innerModel.KeyMap.CursorDown = delegateKeys.cursorDown

	// Additional key mappings for the short and full help views. This allows
	// you to add additional key mappings to the help menu without
	// re-implementing the help component.
	m.innerModel.AdditionalShortHelpKeys = delegateKeys.ShortHelp
	m.innerModel.AdditionalFullHelpKeys = delegateKeys.FullHelp

	m.innerModel.Title = defaultListTitle
	m.innerModel.SetShowStatusBar(false)

	return &m
}

func (m *listModel) Init() tea.Cmd {
	// This function is called from init_$PLATFORM.go file
	return message.TeaCmd(MsgRefreshRepo{})
}

func (m *listModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m, m.handleKeyboardEvent(msg)
	case tea.WindowSizeMsg:
		// Triggers immediately after app start because we render this component by default
		h, v := docStyle.GetFrameSize()
		m.innerModel.SetSize(msg.Width-h, msg.Height-v)
		m.logger.Debug("[UI] Set host list size: %d %d", m.innerModel.Width(), m.innerModel.Height())
		return m, nil
	case MsgRefreshRepo:
		m.logger.Debug("[UI] Load hostnames from the database")
		return m, m.refreshRepo(msg)
	case msgRefreshUI:
		cmd := m.onFocusChanged(msg)
		m.listTitleUpdate()
		m.updateKeyMap()
		return m, cmd
	default:
		return m, m.updateChildModel(msg)
	}
}

func (m *listModel) handleKeyboardEvent(msg tea.KeyMsg) tea.Cmd {
	switch {
	case m.innerModel.SettingFilter():
		m.logger.Debug("[UI] Process key message when in filter mode")
		// If filter is enabled, we should not handle any keyboard messages,
		// as it should be done by filter component.

		// However, there is one special case, which should be taken into account:
		// When user filters out values and presses down key on her keyboard
		// we need to ensure that the title contains proper selection.
		// that's why we need to invoke title update function.
		// See https://github.com/grafviktor/goto/issues/37
		m.listTitleUpdate()

		return m.updateChildModel(msg)
	case m.mode != modeDefault:
		// Handle key event when some mode is enabled. For instance "removeMode".
		return m.handleKeyEventWhenModeEnabled(msg)
	case key.Matches(msg, m.keyMap.connect):
		return m.constructProcessCmd(msg)
	case key.Matches(msg, m.keyMap.remove):
		return m.enterRemoveItemMode()
	case key.Matches(msg, m.keyMap.edit):
		return m.editItem(msg)
	case key.Matches(msg, m.keyMap.append):
		return message.TeaCmd(OpenEditForm{}) // When create a new item, jump to edit mode.
	case key.Matches(msg, m.keyMap.clone):
		return m.copyItem(msg)
	default:
		// If we could not find our own update handler, we pass message to the child model
		// otherwise we would have to implement all key handlers and other stuff by ourselves

		// Dispatch 2 messages:
		// 1 - message which was returned from the inner model.
		// 2 - msgRefreshUI message to update list title. We only need to dispatch it when we switch between list items.
		return tea.Batch(m.updateChildModel(msg), message.TeaCmd(msgRefreshUI{}))
	}
}

func (m *listModel) View() string {
	return docStyle.Render(m.innerModel.View())
}

func (m *listModel) updateChildModel(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	m.innerModel, cmd = m.innerModel.Update(msg)

	return cmd
}

func (m *listModel) updateKeyMap() {
	shouldShowEditButtons := m.innerModel.SelectedItem() != nil

	if shouldShowEditButtons != m.keyMap.ShouldShowEditButtons() {
		m.logger.Debug("[UI] Show edit keyboard shortcuts: %v", shouldShowEditButtons)
		m.keyMap.SetShouldShowEditButtons(shouldShowEditButtons)
	}
}

func (m *listModel) handleKeyEventWhenModeEnabled(msg tea.KeyMsg) tea.Cmd {
	if key.Matches(msg, m.keyMap.confirm) {
		return m.confirmAction()
	}

	// If user doesn't confirm the operation, we go back to normal mode and update
	// title back to normal, this exact key event won't be handled
	m.logger.Debug("[UI] Exit %s mode. Cancel action", m.mode)
	m.mode = modeDefault
	return message.TeaCmd(msgRefreshUI{})
}

func (m *listModel) confirmAction() tea.Cmd {
	if m.mode == modeRemoveItem {
		m.logger.Debug("[UI] Exit %s mode. Confirm action", m.mode)
		m.mode = modeDefault
		return m.removeItem()
	}

	return nil
}

func (m *listModel) enterRemoveItemMode() tea.Cmd {
	// Check if item is selected.
	_, ok := m.innerModel.SelectedItem().(ListItemHost)
	if !ok {
		m.logger.Debug("[UI] Cannot remove. Item is not selected")
		return message.TeaCmd(msgErrorOccurred{err: errors.New(itemNotSelectedMessage)})
	}

	m.mode = modeRemoveItem
	m.logger.Debug("[UI] Enter %s mode. Ask user for confirmation", m.mode)

	return message.TeaCmd(msgRefreshUI{})
}

func (m *listModel) removeItem() tea.Cmd {
	m.logger.Debug("[UI] Remove host from the database")
	item, ok := m.innerModel.SelectedItem().(ListItemHost)
	if !ok {
		m.logger.Error("[UI] Cannot cast selected item to host model")
		return message.TeaCmd(msgErrorOccurred{err: errors.New(itemNotSelectedMessage)})
	}

	err := m.repo.Delete(item.ID)
	if err != nil {
		m.logger.Debug("[UI] Error removing host from the database. %v", err)
		return message.TeaCmd(msgErrorOccurred{err})
	}

	return tea.Batch(
		message.TeaCmd(MsgRefreshRepo{}),
		message.TeaCmd(msgRefreshUI{}),
	)
}

func (m *listModel) refreshRepo(_ tea.Msg) tea.Cmd {
	hosts, err := m.repo.GetAll()
	if err != nil {
		m.logger.Error("[UI] Cannot read database. %v", err)
		return message.TeaCmd(msgErrorOccurred{err})
	}

	slices.SortFunc(hosts, func(a, b model.Host) int {
		if a.Title < b.Title {
			return -1
		}
		return 1
	})

	// Wrap hosts into List items
	items := make([]list.Item, 0, len(hosts))
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

	return tea.Batch(setItemsCmd, message.TeaCmd(msgRefreshUI{}))
}

func (m *listModel) editItem(_ tea.Msg) tea.Cmd {
	item, ok := m.innerModel.SelectedItem().(ListItemHost)
	if !ok {
		return message.TeaCmd(msgErrorOccurred{err: errors.New(itemNotSelectedMessage)})
	}

	host := *item.Unwrap()
	m.logger.Info("[UI] Edit item id: %d, title: %s", host.ID, host.Title)
	return message.TeaCmd(OpenEditForm{HostID: host.ID})
}

func (m *listModel) copyItem(_ tea.Msg) tea.Cmd {
	item, ok := m.innerModel.SelectedItem().(ListItemHost)
	if !ok {
		m.logger.Error("[UI] Cannot cast selected item to host model")
		return message.TeaCmd(msgErrorOccurred{err: errors.New(itemNotSelectedMessage)})
	}

	originalHost := item.Unwrap()
	m.logger.Info("[UI] Copy host item id: %d, title: %s", originalHost.ID, originalHost.Title)
	clonedHost := originalHost.Clone()
	for i := 1; ok; i++ {
		clonedHostTitle := fmt.Sprintf("%s (%d)", originalHost.Title, i)
		listItems := m.innerModel.Items()
		idx := slices.IndexFunc(listItems, func(li list.Item) bool {
			return li.(ListItemHost).Title() == clonedHostTitle
		})

		if idx < 0 {
			clonedHost.Title = clonedHostTitle
			break
		}
	}

	if _, err := m.repo.Save(clonedHost); err != nil {
		return message.TeaCmd(msgErrorOccurred{err})
	}

	return tea.Batch(
		message.TeaCmd(MsgRefreshRepo{}),
		message.TeaCmd(msgRefreshUI{}),
	)
}

func (m *listModel) constructProcessCmd(_ tea.KeyMsg) tea.Cmd {
	item, ok := m.innerModel.SelectedItem().(ListItemHost)
	if !ok {
		m.logger.Error("[UI] Cannot cast selected item to host model")
		return message.TeaCmd(msgErrorOccurred{err: errors.New(itemNotSelectedMessage)})
	}

	return message.TeaCmd(message.RunProcessConnectSSH{Host: *item.Unwrap()})
}

func (m *listModel) listTitleUpdate() {
	var newTitle string

	item, ok := m.innerModel.SelectedItem().(ListItemHost)

	switch {
	case !ok:
		newTitle = defaultListTitle
	case m.mode == modeRemoveItem:
		newTitle = fmt.Sprintf("delete \"%s\" ? (y/N)", item.Title())
	default:
		// Replace Windows ssh prefix "cmd /c ssh" with "ssh"
		newTitle = strings.Replace(item.Unwrap().CmdSSHConnect(), "cmd /c ", "", 1)
		// newTitle = utils.RemoveDuplicateSpaces(newTitle)
	}

	if m.innerModel.Title != newTitle {
		m.innerModel.Title = newTitle
		m.logger.Debug("[UI] New list title: %s", m.innerModel.Title)
	}
}

func (m *listModel) onFocusChanged(_ tea.Msg) tea.Cmd {
	if m.innerModel.SelectedItem() == nil {
		return nil
	}

	if hostItem, ok := m.innerModel.SelectedItem().(ListItemHost); ok {
		m.logger.Debug("[UI] Select host id: %v, title: %s", hostItem.ID, hostItem.Title())
		return tea.Batch(
			message.TeaCmd(message.HostListSelectItem{HostID: hostItem.ID}),
			message.TeaCmd(message.RunProcessLoadSSHConfig{Host: *hostItem.Unwrap()}),
		)
	}

	m.logger.Error("[UI] Select unknown item type from the list")
	return nil
}
