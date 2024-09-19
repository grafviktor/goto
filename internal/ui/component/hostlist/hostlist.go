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
	"github.com/samber/lo"

	"github.com/grafviktor/goto/internal/constant"
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
	OpenEditForm     struct{ HostID int }
	msgErrorOccurred struct{ err error }
	msgToggleLayout  struct{}
)

type listModel struct {
	list.Model
	repo     storage.HostStorage
	keyMap   *keyMap
	appState *state.ApplicationState
	logger   iLogger
	mode     string
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
	m.logger.Debug("[UI] Load hostnames from the database")
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
	selectHostByIDCmd := m.selectHostByID(m.appState.Selected)
	return tea.Sequence(setItemsCmd, selectHostByIDCmd)
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
	case message.HostSSHConfigLoaded:
		m.onHostSSHConfigLoaded(msg)
		return m, nil
	case message.HostUpdated:
		cmd := m.onHostUpdated(msg)
		return m, cmd
	case message.HostCreated:
		cmd := m.onHostCreated(msg)
		return m, cmd
	default:
		return m, m.updateChildModel(msg)
	}
}

func (m *listModel) handleKeyboardEvent(msg tea.KeyMsg) tea.Cmd {
	switch {
	case m.SettingFilter():
		m.logger.Debug("[UI] Process key message when in filter mode")
		// If filter is enabled, we should not handle any keyboard messages,
		// as it should be done by filter component.
		if key.Matches(msg, m.KeyMap.AcceptWhileFiltering) {
			// When user presses enter key while in filter mode, we should load SSH config.
			m.logger.Debug("[UI] Focus item while in filter mode")
			return tea.Sequence(m.updateChildModel(msg), m.onFocusChanged())
		}
		return m.updateChildModel(msg)
	case key.Matches(msg, m.KeyMap.CancelWhileFiltering):
		selectedID := m.SelectedItem().(ListItemHost).ID
		return tea.Sequence(m.updateChildModel(msg), m.selectHostByID(selectedID))
	case key.Matches(msg, m.Model.KeyMap.ClearFilter):
		// When user clears the host filter, child model resets the focus. Explicitly set focus on previously selected item.
		selectedID := m.SelectedItem().(ListItemHost).ID
		return tea.Sequence(m.updateChildModel(msg), m.selectHostByID(selectedID))
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
		return m.editItem()
	case key.Matches(msg, m.keyMap.append):
		return message.TeaCmd(OpenEditForm{}) // When create a new item, jump to edit mode.
	case key.Matches(msg, m.keyMap.clone):
		return m.copyItem()
	case key.Matches(msg, m.keyMap.toggleLayout):
		m.updateChildModel(msgToggleLayout{})
		// When switch between screen layouts, it's required to update pagination.
		// ListModel's updatePagination method is private and cannot be called from
		// here. One of the ways to trigger it is to invoke model.SetSize method.
		m.Model.SetSize(m.Width(), m.Height())
		return nil
	default:
		cmd := m.updateChildModel(msg)
		return tea.Sequence(cmd, m.onFocusChanged())
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

/*
 * Actions.
 */

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

	index := m.Index()
	// If we remove the last item, then we should select the previous one. However, this won't work if only one item is
	// left on the page, because m.VisibleItems() returns nothing. Need to improve.
	isLastPosition := index == len(m.VisibleItems())-1
	m.Model.RemoveItem(index)

	if isLastPosition {
		m.Select(index - 1)
	}

	if item, ok := m.Model.SelectedItem().(ListItemHost); ok {
		return tea.Sequence(
			message.TeaCmd(message.HostListSelectItem{HostID: item.ID}),
			message.TeaCmd(message.RunProcessSSHLoadConfig{Host: item.Host}),
		)
	}

	return nil
}

func (m *listModel) editItem() tea.Cmd {
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

func (m *listModel) copyItem() tea.Cmd {
	item, ok := m.SelectedItem().(ListItemHost)
	if !ok {
		m.logger.Error("[UI] Cannot cast selected item to host model")
		return message.TeaCmd(msgErrorOccurred{err: errors.New(itemNotSelectedMessage)})
	}

	originalHost := item.Host
	m.logger.Info("[UI] Copy host item id: %d, title: %s", originalHost.ID, originalHost.Title)
	clonedHost := originalHost.Clone()
	for i := 1; ok; i++ {
		// Keep generating new title until it's unique
		clonedHostTitle := fmt.Sprintf("%s (%d)", originalHost.Title, i)
		listItems := m.Items()
		idx := slices.IndexFunc(listItems, func(li list.Item) bool {
			return li.(ListItemHost).Title() == clonedHostTitle
		})

		// If title is unique, then we assign the title to the cloned host
		if idx < 0 {
			clonedHost.Title = clonedHostTitle
			break
		}
	}

	var err error
	// Re-assign clonedHost to obtain host ID which is assigned by the database
	if clonedHost, err = m.repo.Save(clonedHost); err != nil {
		return message.TeaCmd(msgErrorOccurred{err})
	}

	titles := lo.Reduce(m.Items(), func(agg []string, item list.Item, index int) []string {
		return append(agg, item.(ListItemHost).Title())
	}, []string{clonedHost.Title})

	slices.Sort(titles)
	index := lo.IndexOf(titles, clonedHost.Title)
	return m.Model.InsertItem(index, ListItemHost{Host: clonedHost})
}

/*
 * Event handlers - those events come from other components.
 */

func (m *listModel) onHostUpdated(msg message.HostUpdated) tea.Cmd {
	var cmd tea.Cmd
	updatedItem := ListItemHost{Host: msg.Host}
	titles := lo.Map(m.Items(), func(item list.Item, index int) string {
		if item.(ListItemHost).ID == updatedItem.ID {
			return updatedItem.Title()
		}

		return item.(ListItemHost).Title()
	})

	slices.Sort(titles)
	newIndex := lo.IndexOf(titles, updatedItem.Title())

	if newIndex == m.Index() {
		// Index isn't changed.
		cmd = m.Model.SetItem(m.Index(), updatedItem)
	} else {
		// Index is changed, need to move the host into a new location
		m.Model.RemoveItem(m.Index())
		cmd = m.Model.InsertItem(newIndex, updatedItem)
		m.Select(newIndex)
	}

	m.updateTitle()

	return tea.Sequence(
		cmd,
		message.TeaCmd(message.HostListSelectItem{HostID: msg.Host.ID}),
		// See S1016 - Use a type conversion instead of manually copying struct fields one by one.
		message.TeaCmd(message.RunProcessSSHLoadConfig(msg)),
	)
}

func (m *listModel) onHostCreated(msg message.HostCreated) tea.Cmd {
	listItem := ListItemHost{Host: msg.Host}
	titles := lo.Reduce(m.Items(), func(agg []string, item list.Item, index int) []string {
		return append(agg, item.(ListItemHost).Title())
	}, []string{listItem.Title()})

	slices.Sort(titles)
	index := lo.IndexOf(titles, listItem.Title())
	cmd := m.Model.InsertItem(index, listItem)

	m.Select(index)
	m.updateTitle()

	return tea.Sequence(
		// If host position coincides with other host, then let the underlying model to handle that
		cmd,
		message.TeaCmd(message.HostListSelectItem{HostID: msg.Host.ID}),
		// See S1016 - Use a type conversion instead of manually copying struct fields one by one.
		message.TeaCmd(message.RunProcessSSHLoadConfig(msg)),
	)
}

func (m *listModel) onFocusChanged() tea.Cmd {
	if m.SelectedItem() == nil {
		m.logger.Debug("[UI] Focus is not set to any item in the list")
	}

	if hostItem, ok := m.SelectedItem().(ListItemHost); ok {
		m.logger.Debug("[UI] Focus changed to host id: %v, title: %s", hostItem.ID, hostItem.Title())
		m.updateTitle()
		m.updateKeyMap()

		return tea.Sequence(
			message.TeaCmd(message.HostListSelectItem{HostID: hostItem.ID}),
			message.TeaCmd(message.RunProcessSSHLoadConfig{Host: hostItem.Host}),
		)
	}

	m.logger.Error("[UI] Select unknown item type from the list")
	return nil
}

func (m *listModel) onHostSSHConfigLoaded(msg message.HostSSHConfigLoaded) {
	for index, item := range m.Items() {
		if hostListItem, ok := item.(ListItemHost); ok && hostListItem.ID == msg.HostID {
			hostListItem.SSHClientConfig = &msg.Config
			m.SetItem(index, hostListItem)
			break
		}
	}
}

/*
 * Helper methods.
 */

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

func (m *listModel) updateTitle() {
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

func (m *listModel) selectHostByID(id int) tea.Cmd {
	_, index, found := lo.FindIndexOf(m.Items(), func(item list.Item) bool {
		hostItem, ok := item.(ListItemHost)
		return ok && hostItem.ID == id
	})

	if found {
		m.Select(index)
		return m.onFocusChanged()
	}

	return nil
}

/*
 * Deal with actions which require confirmation from the user.
 */

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
	m.updateTitle()

	return nil
}

func (m *listModel) handleKeyEventWhenModeEnabled(msg tea.KeyMsg) tea.Cmd {
	if key.Matches(msg, m.keyMap.confirm) {
		return m.confirmAction()
	}

	// If user doesn't confirm the operation, we go back to normal mode and update
	// title back to normal, this exact key event won't be handled
	m.logger.Debug("[UI] Exit %s mode. Cancel action.", m.mode)
	m.mode = modeDefault
	return m.selectHostByID(m.appState.Selected)
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
	m.updateTitle()

	return cmd
}
