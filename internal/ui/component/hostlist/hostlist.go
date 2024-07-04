// Package hostlist implements the host list view.
package hostlist

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/exp/slices"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	hostModel "github.com/grafviktor/goto/internal/model/host"
	"github.com/grafviktor/goto/internal/state"
	"github.com/grafviktor/goto/internal/storage"
	"github.com/grafviktor/goto/internal/ui/message"
	"github.com/grafviktor/goto/internal/utils"
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
	MsgRefreshRepo  struct{}
	msgRefreshUI    struct{}
	msgToggleLayout struct{}
)

type listModel struct {
	list.Model
	repo     storage.HostStorage
	keyMap   *keyMap
	appState *state.ApplicationState
	logger   iLogger
	mode     string
	// That is a small optimization, as we do not want to re-read host configuration
	// every time when we dispatch msgRefreshUI{} message.
	prevSelectedItemID int
}

// New - creates new host list model.
// context - is not used.
// storage - is the data layer.
// appState - is the application state, usually we want to restore previous state when application restarts,
// for instance focus previously selected host.
// log - application logger.
func New(_ context.Context, storage storage.HostStorage, appState *state.ApplicationState, log iLogger) *listModel {
	// delegate := buildScreenLayout(appState.ScreenLayout)
	delegate := NewHostDelegate(&appState.ScreenLayout, log)
	delegateKeys := newDelegateKeyMap()

	var listItems []list.Item
	model := list.New(listItems, delegate, 0, 0)
	// This line affects sorting when filtering enabled. What UnsortedFilter
	// does - it filters the collection, but leaves initial items order unchanged.
	// Default filter on the contrary - filters the collection based on the match rank.
	model.Filter = list.UnsortedFilter

	m := listModel{
		Model:    model,
		keyMap:   delegateKeys,
		repo:     storage,
		appState: appState,
		logger:   log,
	}

	m.KeyMap.CursorUp.Unbind()
	m.KeyMap.CursorUp = delegateKeys.cursorUp
	m.KeyMap.CursorDown.Unbind()
	m.KeyMap.CursorDown = delegateKeys.cursorDown

	// Additional key mappings for the short and full help views. This allows
	// you to add additional key mappings to the help menu without
	// re-implementing the help component.
	m.AdditionalShortHelpKeys = delegateKeys.ShortHelp
	m.AdditionalFullHelpKeys = delegateKeys.FullHelp

	m.Title = defaultListTitle
	m.SetShowStatusBar(false)

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
		m.SetSize(msg.Width-h, msg.Height-v)
		m.logger.Debug("[UI] Set host list size: %d %d", m.Width(), m.Height())
		return m, nil
	case MsgRefreshRepo:
		m.logger.Debug("[UI] Load hostnames from the database")
		return m, m.refreshRepo(msg)
	case msgRefreshUI:
		return m, m.onFocusChanged()
	default:
		return m, m.updateChildModel(msg)
	}
}

func (m *listModel) handleKeyboardEvent(msg tea.KeyMsg) tea.Cmd {
	switch {
	case key.Matches(msg, m.KeyMap.AcceptWhileFiltering):
		m.logger.Debug("[UI] Focus item while in filter mode")
		return tea.Batch(m.updateChildModel(msg), m.onFocusChanged())
	case m.SettingFilter():
		m.logger.Debug("[UI] Process key message when in filter mode")
		// If filter is enabled, we should not handle any keyboard messages,
		// as it should be done by filter component.
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
	case key.Matches(msg, m.keyMap.toggleLayout):
		m.updateChildModel(msgToggleLayout{})
		// When switch between screen layouts, it's required to update pagination.
		// ListModel's updatePagination method is private and cannot be called from
		// here. One of the ways to trigger it is to invoke model.SetSize method.
		m.Model.SetSize(m.Width(), m.Height())

		return nil
	case key.Matches(msg, m.Model.KeyMap.ClearFilter):
		// When user clears the host filter, keep the focus on the selected item.
		cmd := m.updateChildModel(msg)
		m.selectItemByModelId(m.prevSelectedItemID)
		return cmd
	default:
		// If we could not find our own update handler, we pass message to the child model
		// otherwise we would have to implement all key handlers and other stuff by ourselves

		// Dispatch several messages:
		// 1 - message which was returned from the inner model.
		// 2 - messages which returned by onFocusChanged, which will trigger the SSH configuration load.
		return tea.Sequence(m.updateChildModel(msg), m.onFocusChanged())
	}
}

func (m *listModel) View() string {
	// BUG: On a certain screen width 'docStyle.Render(...)' adds an extra line
	// break symbol when a host item title contains '-' symbol. That's most
	// probably the bubbletea library bug, but still requires analysis.
	return docStyle.Render(m.Model.View())
}

func (m *listModel) updateChildModel(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	m.Model, cmd = m.Model.Update(msg)

	return cmd
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
	_, ok := m.SelectedItem().(ListItemHost)
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
	item, ok := m.SelectedItem().(ListItemHost)
	if !ok {
		m.logger.Error("[UI] Cannot cast selected item to host model")
		return message.TeaCmd(msgErrorOccurred{err: errors.New(itemNotSelectedMessage)})
	}

	err := m.repo.Delete(item.ID)
	if err != nil {
		m.logger.Debug("[UI] Error removing host from the database. %v", err)
		return message.TeaCmd(msgErrorOccurred{err})
	}

	// That's a hack! When we delete an item, the inner model automatically changes focus to an existing item
	// without sending any notification. If we do not reset the prevSelectedItemID, the onFocusChanged function
	// will not be triggered, and we will not load the SSH configuration for the new selected item.
	// Probably it's worth to explicitly focus a new item after deletion.
	m.prevSelectedItemID = -1

	// This should be replaced with tea.Sequence as msgRefreshUI completes before MsgRefreshRepo, as a result,
	// the application reads a configuration of a host which was just deleted. This bug only appears when there is
	// only one host left in the database and we delete it.
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

	slices.SortFunc(hosts, func(a, b hostModel.Host) int {
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

	setItemsCmd := m.SetItems(items)

	return tea.Sequence(setItemsCmd, message.TeaCmd(msgRefreshUI{}))
}

func (m *listModel) editItem(_ tea.Msg) tea.Cmd {
	item, ok := m.SelectedItem().(ListItemHost)
	if !ok {
		return message.TeaCmd(msgErrorOccurred{err: errors.New(itemNotSelectedMessage)})
	}

	m.logger.Info("[UI] Edit item id: %d, title: %s", item.ID, item.Title())
	return tea.Sequence(
		message.TeaCmd(OpenEditForm{HostID: item.ID}),
		// Load SSH config for the selected host
		message.TeaCmd(message.RunProcessLoadSSHConfig{Host: item.Host}),
	)
}

func (m *listModel) copyItem(_ tea.Msg) tea.Cmd {
	item, ok := m.SelectedItem().(ListItemHost)
	if !ok {
		m.logger.Error("[UI] Cannot cast selected item to host model")
		return message.TeaCmd(msgErrorOccurred{err: errors.New(itemNotSelectedMessage)})
	}

	originalHost := item.Host
	m.logger.Info("[UI] Copy host item id: %d, title: %s", originalHost.ID, originalHost.Title)
	clonedHost := originalHost.Clone()
	for i := 1; ok; i++ {
		clonedHostTitle := fmt.Sprintf("%s (%d)", originalHost.Title, i)
		listItems := m.Items()
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

	// Do not need to dispatch msgRefreshUI{} here as onFocus change event will trigger anyway
	return message.TeaCmd(MsgRefreshRepo{})
}

func (m *listModel) constructProcessCmd(_ tea.KeyMsg) tea.Cmd {
	item, ok := m.SelectedItem().(ListItemHost)
	if !ok {
		m.logger.Error("[UI] Cannot cast selected item to host model")
		return message.TeaCmd(msgErrorOccurred{err: errors.New(itemNotSelectedMessage)})
	}

	return message.TeaCmd(message.RunProcessConnectSSH{Host: item.Host})
}

func (m *listModel) onFocusChanged() tea.Cmd {
	// BUG: when create a new host, the focus is not set to the new item.
	// because onFocusChanged() redefines the focus to the previous item.
	m.listTitleUpdate()
	m.updateKeyMap()

	if m.SelectedItem() == nil {
		m.logger.Debug("[UI] Focus is not set to any item in the list")
		// Here we can set the default focus to the first item in the list.
		return nil
	}

	if hostItem, ok := m.SelectedItem().(ListItemHost); ok {
		m.logger.Debug("[UI] Check if selection changed. Prev item: %v, Curr item: %v", m.prevSelectedItemID, hostItem.ID)
		if m.prevSelectedItemID != hostItem.ID {
			m.prevSelectedItemID = hostItem.ID
			m.logger.Debug("[UI] Focus changed to host id: %v, title: %s", hostItem.ID, hostItem.Title())
			return tea.Batch(
				message.TeaCmd(message.HostListSelectItem{HostID: hostItem.ID}),
				message.TeaCmd(message.RunProcessLoadSSHConfig{Host: hostItem.Host}),
			)
		}
	} else {
		m.logger.Error("[UI] Select unknown item type from the list")
	}

	return nil
}

func (m *listModel) listTitleUpdate() {
	var newTitle string

	item, ok := m.SelectedItem().(ListItemHost)

	switch {
	case !ok:
		newTitle = defaultListTitle
	case m.mode == modeRemoveItem:
		newTitle = fmt.Sprintf("delete \"%s\" ? (y/N)", item.Title())
	default:
		// Replace Windows ssh prefix "cmd /c ssh" with "ssh"
		newTitle = strings.Replace(item.Host.CmdSSHConnect(), "cmd /c ", "", 1)
		newTitle = utils.RemoveDuplicateSpaces(newTitle)
	}

	if m.Title != newTitle {
		m.Title = newTitle
		m.logger.Debug("[UI] New list title: %s", m.Title)
	}
}

func (m *listModel) updateKeyMap() {
	shouldShowEditButtons := m.SelectedItem() != nil

	if shouldShowEditButtons != m.keyMap.ShouldShowEditButtons() {
		m.logger.Debug("[UI] Show edit keyboard shortcuts: %v", shouldShowEditButtons)
		m.keyMap.SetShouldShowEditButtons(shouldShowEditButtons)
	}
}

func (m *listModel) selectItemByModelId(id int) {
	for i, item := range m.VisibleItems() {
		if hostItem, ok := item.(ListItemHost); ok {
			if hostItem.ID == id {
				m.Select(i)
				break
			}
		}
	}
}
