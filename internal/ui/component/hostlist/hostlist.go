// Package hostlist implements the host list view.
package hostlist

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/grafviktor/goto/internal/constant"

	"golang.org/x/exp/slices"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/samber/lo"

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
	modeSSHCopyID          = "sshCopyID"
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
		m.selectItemByModelID(m.appState.Selected)
		return m, m.onFocusChanged()
	case message.HostSSHConfigLoaded:
		m.handleHostSSHConfigLoaded(msg)
		return m, nil
	case message.HostUpdated:
		// FIXME: Should update title
		cmd := m.handleHostUpdated(msg)
		return m, cmd
	case message.HostCreated:
		// FIXME: Should update title
		cmd := m.handleHostCreated(msg)
		return m, cmd
	default:
		return m, m.updateChildModel(msg)
	}
}

func (m *listModel) handleHostUpdated(msg message.HostUpdated) tea.Cmd {
	var cmds []tea.Cmd
	listItem := ListItemHost{Host: msg.Host}
	titles := lo.Map(m.Items(), func(item list.Item, index int) string {
		if item.(ListItemHost).ID == listItem.ID {
			return listItem.Title()
		}

		return item.(ListItemHost).Title()
	})

	slices.Sort(titles)
	newIndex := lo.IndexOf(titles, listItem.Title())

	if newIndex == m.Index() {
		cmds = append(cmds, m.Model.SetItem(m.Index(), listItem))
	} else {
		m.Model.RemoveItem(m.Index())
		cmds = append(cmds, m.Model.InsertItem(newIndex, listItem))
		m.Select(newIndex)
	}

	// cmds = append(cmds, m.Model.SetItem(m.Index(), listItem))

	m.listTitleUpdate()
	cmds = append(cmds, message.TeaCmd(message.HostListSelectItem{HostID: msg.Host.ID}))
	cmds = append(cmds, message.TeaCmd(message.RunProcessSSHLoadConfig{Host: msg.Host}))

	return tea.Sequence(cmds...)

	// if newIndex == m.Index() {
	// 	// If host position isn't changed, then just replace it in hostlist using the same index
	// 	cmds = append(cmds, m.Model.SetItem(m.Index(), listItem))
	// 	m.listTitleUpdate()
	// } else if newIndex == len(m.Model.Items()) {
	// 	// If host position is last, then remove it from current position and place it in the end of collection
	// 	m.Model.RemoveItem(m.Index())
	// 	cmds = append(cmds, m.Model.InsertItem(newIndex-1, listItem))
	// 	m.Select(newIndex - 1)
	// 	m.listTitleUpdate()
	// } else {
	// 	// If host position coincides with other host, then shift other host to the new place
	// 	temp := m.Model.Items()[newIndex].(ListItemHost)
	// 	cmds = append(cmds, m.Model.SetItem(newIndex, listItem))
	// 	m.Select(newIndex)
	// 	m.listTitleUpdate()
	// 	// Because we're moving the host to a new place, we should remove
	// 	// it from existing one. If item does not exist, there will be no error
	// 	m.Model.RemoveItem(m.Index()) // m.Index() contains current host position
	// 	newIndex = m.indexOfHost(temp)
	// 	// Is this a bug? Probably not, because we use InsertItem, not SetItem
	// 	// Because we're inserting the temprorary item at a specific index where
	// 	// another host can be. It's required to run a recursive call here
	// 	cmds = append(cmds, m.Model.InsertItem(newIndex, temp))
	// }

	// cmds = append(cmds, message.TeaCmd(message.HostListSelectItem{HostID: msg.Host.ID}))
	// cmds = append(cmds, message.TeaCmd(message.RunProcessSSHLoadConfig{Host: msg.Host}))

	// // return cmd
	// return tea.Sequence(cmds...)
}

func (m *listModel) handleHostCreated(msg message.HostCreated) tea.Cmd {
	listItem := ListItemHost{Host: msg.Host}
	// FIXME: title cotains duplicate item
	titles := lo.Reduce(m.Items(), func(agg []string, item list.Item, index int) []string {
		return append(agg, item.(ListItemHost).Title())
	}, []string{listItem.Title()})

	slices.Sort(append(titles, listItem.Title()))
	index := lo.IndexOf(titles, listItem.Title())

	m.Select(index)
	m.listTitleUpdate()

	cmd := m.Model.InsertItem(index, listItem)
	return tea.Sequence(
		// If host position coincides with other host, then let the underlying model to handle that
		cmd,
		message.TeaCmd(message.HostListSelectItem{HostID: msg.Host.ID}),
		message.TeaCmd(message.RunProcessSSHLoadConfig{Host: msg.Host}),
	)
}

// func (m *listModel) indexOfHost(host ListItemHost, isNewHost bool) int {
// 	titles := lo.Map(m.Items(), func(i list.Item, index int) string {
// 		return i.(ListItemHost).Title()
// 	})

// 	if isNewHost {
// 		titles = append(titles, host.Title())
// 	}

// 	slices.Sort(titles)
// 	return lo.IndexOf(titles, host.Title())
// }

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
		return m.constructProcessCmd(constant.ProcessTypeSSHConnect)
	case key.Matches(msg, m.keyMap.copyID):
		return m.enterSSHCopyIDMode()
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
		m.selectItemByModelID(m.prevSelectedItemID)
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
	m.logger.Debug("[UI] Exit %s mode. Cancel action.", m.mode)
	m.mode = modeDefault
	return message.TeaCmd(msgRefreshUI{})
}

func (m *listModel) confirmAction() tea.Cmd {
	var cmd tea.Cmd
	if m.mode == modeRemoveItem {
		cmd = m.removeItem()
	} else if m.mode == modeSSHCopyID {
		cmd = m.constructProcessCmd(constant.ProcessTypeSSHCopyID)
	}

	m.logger.Debug("[UI] Exit %s mode. Confirm action.", m.mode)
	m.mode = modeDefault
	m.listTitleUpdate()

	return cmd
}

func (m *listModel) enterSSHCopyIDMode() tea.Cmd {
	// Check if item is selected.
	_, ok := m.SelectedItem().(ListItemHost)
	if !ok {
		m.logger.Debug("[UI] Cannot copy id. Host is not selected.")
		return message.TeaCmd(msgErrorOccurred{err: errors.New(itemNotSelectedMessage)})
	}

	m.mode = modeSSHCopyID
	m.logger.Debug("[UI] Enter %s mode. Ask user for confirmation.", m.mode)
	m.Title = "copy ssh key to the remote host? (y/N)"

	return nil
}

func (m *listModel) enterRemoveItemMode() tea.Cmd {
	// Check if item is selected.
	_, ok := m.SelectedItem().(ListItemHost)
	if !ok {
		m.logger.Debug("[UI] Cannot remove. Host is not selected.")
		return message.TeaCmd(msgErrorOccurred{err: errors.New(itemNotSelectedMessage)})
	}

	m.mode = modeRemoveItem
	m.logger.Debug("[UI] Enter %s mode. Ask user for confirmation.", m.mode)

	// Ideally, we should not return msgRefreshUI{} from this function,
	// but title is not getting updated. Requires investigation.
	return message.TeaCmd(msgRefreshUI{})
}

func (m *listModel) removeItem() tea.Cmd {
	m.logger.Debug("[UI] Remove host from the database")
	item, ok := m.SelectedItem().(ListItemHost)
	if !ok {
		// We should not be here at all, because delete
		// button isn't available when a host is not selected.
		m.logger.Error("[UI] Cannot cast selected item to host model")
		return message.TeaCmd(msgErrorOccurred{err: errors.New(itemNotSelectedMessage)})
	}

	err := m.repo.Delete(item.ID)
	if err != nil {
		m.logger.Debug("[UI] Error removing host from the database. %v", err)
		return message.TeaCmd(msgErrorOccurred{err})
	}

	// To check,
	// 1. cmd := m.Model.RemoveItem(m.Index())
	// 2. whether it's possible to call m.listTitleUpdate() title without returning msgRefreshUI
	m.Model.RemoveItem(m.Index())
	if item, ok := m.Model.SelectedItem().(ListItemHost); ok {
		return tea.Sequence(
			message.TeaCmd(message.HostListSelectItem{HostID: item.ID}),
			message.TeaCmd(message.RunProcessSSHLoadConfig{Host: item.Host}),
		)
	}

	return nil
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
		message.TeaCmd(message.RunProcessSSHLoadConfig{Host: item.Host}),
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

	return message.TeaCmd(MsgRefreshRepo{})
}

func (m *listModel) constructProcessCmd(processType constant.ProcessType) tea.Cmd {
	item, ok := m.SelectedItem().(ListItemHost)
	if !ok {
		m.logger.Error("[UI] Cannot cast selected item to host model")
		return message.TeaCmd(msgErrorOccurred{err: errors.New(itemNotSelectedMessage)})
	}

	if processType == constant.ProcessTypeSSHConnect {
		return message.TeaCmd(message.RunProcessSSHConnect{Host: item.Host})
	} else if processType == constant.ProcessTypeSSHCopyID {
		return message.TeaCmd(message.RunProcessSSHCopyID{Host: item.Host})
	}

	return nil
}

func (m *listModel) onFocusChanged() tea.Cmd {
	m.logger.Debug("m.Index(): %d", m.Index())
	m.listTitleUpdate()
	m.updateKeyMap()

	if m.SelectedItem() == nil {
		m.logger.Debug("[UI] Focus is not set to any item in the list")
		return nil
	}

	if hostItem, ok := m.SelectedItem().(ListItemHost); ok {
		m.logger.Debug("[UI] Check if selection changed. Prev item: %v, Curr item: %v", m.prevSelectedItemID, hostItem.ID)
		if m.prevSelectedItemID != hostItem.ID {
			m.prevSelectedItemID = hostItem.ID
			m.logger.Debug("[UI] Focus changed to host id: %v, title: %s", hostItem.ID, hostItem.Title())
			return tea.Batch(
				message.TeaCmd(message.HostListSelectItem{HostID: hostItem.ID}),
				message.TeaCmd(message.RunProcessSSHLoadConfig{Host: hostItem.Host}),
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

func (m *listModel) selectItemByModelID(id int) {
	_, index, found := lo.FindIndexOf(m.VisibleItems(), func(item list.Item) bool {
		hostItem, ok := item.(ListItemHost)
		return ok && hostItem.ID == id
	})

	if found {
		m.Select(index)
	}
}

func (m *listModel) handleHostSSHConfigLoaded(msg message.HostSSHConfigLoaded) {
	if hostListItem, ok := m.SelectedItem().(ListItemHost); ok {
		hostListItem.SSHClientConfig = &msg.Config
		m.SetItem(m.Index(), hostListItem)
	}
}
