// Package hostlist implements the host list view.
package hostlist

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

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
	styleDoc                   = lipgloss.NewStyle().Margin(1, 2, 1, 0)
	itemNotSelectedErrMsg      = "you must select an item"
	modeCloseApp               = "closeApp"
	modeDefault                = ""
	modeRemoveItem             = "removeItem"
	modeSSHCopyID              = "sshCopyID"
	defaultListTitle           = "press 'n' to add a new host"
	notificationMessageTimeout = time.Second * 2
)

type iLogger interface {
	Debug(format string, args ...any)
	Info(format string, args ...any)
	Error(format string, args ...any)
}

type (
	msgToggleLayout     struct{ layout constant.ScreenLayout }
	msgHideNotification struct{}
)

type listModel struct {
	list.Model
	repo                       storage.HostStorage
	keyMap                     *keyMap
	appState                   *state.ApplicationState
	logger                     iLogger
	mode                       string
	notificationMessageTimer   *time.Timer
	notificationMessageTimeout time.Duration
	Styles                     styles
}

// New - creates new host list model.
// context - is not used.
// storage - is the data layer.
// appState - is the application state, usually we want to restore previous state when application restarts,
// for instance focus previously selected host.
// log - application logger.
func New(_ context.Context, storage storage.HostStorage, appState *state.ApplicationState, log iLogger) *listModel {
	delegate := NewHostDelegate(&appState.ScreenLayout, &appState.Group, log)
	delegateKeys := newDelegateKeyMap()
	customStyles := customStyles()

	var listItems []list.Item
	model := list.New(listItems, delegate, 0, 0)
	// This line affects sorting when filtering enabled. What UnsortedFilter
	// does - it filters the collection, but leaves initial items order unchanged.
	// Default filter on the contrary - filters the collection based on the match rank.
	model.Filter = list.UnsortedFilter
	model.Styles = customStyles.Styles

	m := listModel{
		Model:    model,
		keyMap:   delegateKeys,
		repo:     storage,
		appState: appState,
		logger:   log,
		Styles:   customStyles,
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

	m.Title = m.Styles.Title.Render(defaultListTitle)
	m.notificationMessageTimeout = notificationMessageTimeout

	return &m
}

func (m *listModel) Init() tea.Cmd {
	// This function is called from model.go#init() file
	return m.loadHosts()
}

func (m *listModel) loadHosts() tea.Cmd {
	m.logger.Debug("[UI] Load hostnames from the database")
	hosts, err := m.repo.GetAll()
	if err != nil {
		m.logger.Error("[UI] Cannot read database. %v", err)
		return message.TeaCmd(message.ErrorOccurred{Err: err})
	}

	// If host group is selected only load hosts from this group.
	if m.appState.Group != "" {
		hosts = lo.Filter(hosts, func(h hostModel.Host, index int) bool {
			return strings.EqualFold(h.Group, m.appState.Group)
		})
	}

	// Wrap hosts into List items.
	items := make([]list.Item, 0, len(hosts))
	for _, h := range hosts {
		items = append(items, ListItemHost{Host: h})
	}

	// BUG: This sorting is different from one which is used to insert a new host.
	// See sorting in copyItem() method
	sort.Slice(items, func(i, j int) bool {
		uniqueName1 := items[i].(ListItemHost).uniqueName()
		uniqueName2 := items[j].(ListItemHost).uniqueName()
		return uniqueName1 < uniqueName2
	})

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
		h, v := styleDoc.GetFrameSize()
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
	case message.GroupSelected:
		m.logger.Debug("[UI] Update app state. Active group: '%s'", msg.Name)
		m.appState.Group = msg.Name
		// Reset filter when group is selected
		m.ResetFilter()
		// We re-load hosts every time a group is selected. This is not the best way
		// to handle this, as it leads to series of hacks here and there. But it's the
		// simplest way to implement it.
		return m, m.loadHosts()
	case msgHideNotification:
		m.updateTitle()
		return m, nil
	case message.ErrorOccurred:
		return m, m.displayNotificationMsg(msg.Err.Error())
	default:
		m.SetShowStatusBar(m.FilterState() != list.Unfiltered)
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
		} else if key.Matches(msg, m.KeyMap.CancelWhileFiltering) {
			// When user presses Escape key while in filter mode without accepting filter results.
			// Be aware that focus will be set to the first item from the search results, though we haven't
			// selected it explicitly.
			m.logger.Debug("[UI] Clear and exit filter mode")

			if hostItem, ok := m.SelectedItem().(ListItemHost); ok {
				selectedID := hostItem.ID
				return tea.Sequence(m.updateChildModel(msg), m.selectHostByID(selectedID))
			}
		}
		return m.updateChildModel(msg)
	case key.Matches(msg, m.Model.KeyMap.ClearFilter):
		// When user clears the host filter, child model resets the focus. Explicitly set focus on previously selected item.
		selectedID := m.SelectedItem().(ListItemHost).ID
		return tea.Sequence(m.updateChildModel(msg), m.selectHostByID(selectedID))
	case m.mode != modeDefault:
		// Handle key event when some mode is enabled. For instance "removeMode".
		return m.handleKeyEventWhenModeEnabled(msg)
	case key.Matches(msg, m.keyMap.selectGroup):
		return message.TeaCmd(message.OpenViewSelectGroup{})
	case key.Matches(msg, m.keyMap.connect):
		return m.constructProcessCmd(constant.ProcessTypeSSHConnect)
	case key.Matches(msg, m.keyMap.copyID):
		return m.enterSSHCopyIDMode()
	case key.Matches(msg, m.keyMap.remove):
		return m.enterRemoveItemMode()
	case key.Matches(msg, m.keyMap.edit):
		return m.editItem()
	case key.Matches(msg, m.keyMap.append):
		return message.TeaCmd(message.OpenViewHostEdit{}) // When create a new item, jump to edit mode.
	case key.Matches(msg, m.keyMap.clone):
		return m.copyItem()
	case key.Matches(msg, m.keyMap.toggleLayout):
		return m.onToggleLayout()
	case msg.Type == tea.KeyEsc:
		if m.appState.Group != "" {
			// When user presses Escape key while group is selected,
			// we should open select group form.
			m.logger.Debug("[UI] Receive Escape key when group selected. Open view select group.")
			return message.TeaCmd(message.OpenViewSelectGroup{})
		} else {
			m.logger.Debug("[UI] Receive Escape key. Ask user for confirmation to close the app.")
			m.enterCloseAppMode()
			return nil
		}
	default:
		cmd := m.updateChildModel(msg)
		return tea.Sequence(cmd, m.onFocusChanged())
	}
}

func (m *listModel) View() string {
	return styleDoc.Render(m.Model.View())
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
	// Potential bug when filter is enabled as selected item reads from collection duplicate!
	// Consider taking "item" from m.Items()
	item, ok := m.SelectedItem().(ListItemHost)
	if !ok {
		// We should not be here at all, because delete
		// button isn't available when a host is not selected.
		m.logger.Error("[UI] Cannot cast selected item to host model")
		return message.TeaCmd(message.ErrorOccurred{Err: errors.New(itemNotSelectedErrMsg)})
	}

	err := m.repo.Delete(item.ID)
	if err != nil {
		m.logger.Debug("[UI] Error removing host from the database. %v", err)
		return message.TeaCmd(message.ErrorOccurred{Err: err})
	}

	_, index, _ := lo.FindIndexOf(m.Items(), func(i list.Item) bool {
		return i.(ListItemHost).ID == item.ID
	})

	/*
		nolint-godox BUG: Steps to reproduce:
		Create hosts with following titles:
		1
		1 (7)
		1 (7) (1)
		1 (7) (2)
		1 (7) (3)

		Go into filter mode and type "7"
		Delete host "1 (7) (2)"
		Go to host "1 (7)" and copy it
		Host "1 (7) (2)" will be re-created. All looks correct:
		1
		1 (7)
		1 (7) (1)
		1 (7) (2) // If you try to edit it, all fields will be empty
		1 (7) (3)

		Now restart the application and notice that re-created host was saved with a wrong title:
		1
		1 (7)
		1 (7) (1)
		1 (7) (1) // WRONG: Should be "1 (7) (2)"
		1 (7) (3)

		Brief analysis:
		When filter is enabled and user deletes and item, list.go modifies 2 collections simultaneously:
		1. m.items
		2. m.filteredItems

		These 2 collections are modified by RemoveItem method where we send the index of the item which should be removed.
		Because the index of the same item can be different in m.items and m.filteredItems, RemoveItem deletes different
		items in m.Items and m.filteredItems:

		m.items
			1
			1 (7)
			1 (7) (1)
			1 (7) (2) // index = 3, in m.items we delete item "1 (7) (2)"
			1 (7) (3)

		m.filteredItems
			1 (7)
			1 (7) (1)
			1 (7) (2)
			1 (7) (3) // index = 3, in m.filteredItems we delete item "1 (7) (3)"

		To raise a bug in https://github.com/charmbracelet/bubbles project
	*/
	m.Model.RemoveItem(index)
	// We have to reset filter when remove an item from the list because of the aforementioned bug.
	m.Model.ResetFilter()

	if index >= 1 {
		// If it's not the first item in the list, then let's focus on the previous one.
		m.Select(index - 1)
	}

	return tea.Sequence(
		m.onFocusChanged(),
		m.displayNotificationMsg(fmt.Sprintf("deleted \"%s\"", item.Title())),
	)
}

func (m *listModel) editItem() tea.Cmd {
	item, ok := m.SelectedItem().(ListItemHost)
	if !ok {
		return message.TeaCmd(message.ErrorOccurred{Err: errors.New(itemNotSelectedErrMsg)})
	}

	// m.Model.ResetFilter()
	m.logger.Info("[UI] Edit item id: %d, title: %s", item.ID, item.Title())
	return tea.Sequence(
		message.TeaCmd(message.OpenViewHostEdit{HostID: item.ID}),
		// Load SSH config for the selected host
		message.TeaCmd(message.RunProcessSSHLoadConfig{Host: item.Host}),
	)
}

func (m *listModel) copyItem() tea.Cmd {
	item, ok := m.SelectedItem().(ListItemHost)
	if !ok {
		m.logger.Error("[UI] Cannot cast selected item to host model")
		return message.TeaCmd(message.ErrorOccurred{Err: errors.New(itemNotSelectedErrMsg)})
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
		return message.TeaCmd(message.ErrorOccurred{Err: err})
	}

	titles := lo.Reduce(m.Items(), func(agg []string, item list.Item, index int) []string {
		return append(agg, item.(ListItemHost).Title())
	}, []string{clonedHost.Title})

	slices.Sort(titles)
	index := lo.IndexOf(titles, clonedHost.Title)
	// We should NOT call onFocusChanged here, because we do not change focus when copying an item.
	return tea.Sequence(
		m.Model.InsertItem(index, ListItemHost{Host: clonedHost}),
		m.displayNotificationMsg(fmt.Sprintf("cloned \"%s\"", clonedHost.Title)),
	)
}

/*
 * Event handlers - those events come from other components.
 */

// onHostUpdated - not only updates a host, it also re-inserts the host into
// a correct position of the host list, to keep it sorted.
func (m *listModel) onHostUpdated(msg message.HostUpdated) tea.Cmd {
	updatedHost := ListItemHost{Host: msg.Host}
	// Get all item titles, replacing the updated host's title
	allTitles := lo.Map(m.Items(), func(item list.Item, _ int) string {
		host := item.(ListItemHost)
		return lo.Ternary(host.ID == updatedHost.ID, updatedHost.uniqueName(), host.uniqueName())
	})

	slices.Sort(allTitles)
	newIndex := lo.IndexOf(allTitles, updatedHost.uniqueName())

	_, currentIndex, _ := lo.FindIndexOf(m.Items(), func(item list.Item) bool {
		return updatedHost.ID == item.(ListItemHost).ID
	})

	// Do not use m.Index(), as it returns Visible Index, whilst
	// all other functions require the index among all items.
	cmd := lo.Ternary(
		newIndex == currentIndex,
		m.SetItem(currentIndex, updatedHost),
		m.setItemAndReorder(newIndex, currentIndex, updatedHost),
	)

	return tea.Sequence(
		cmd,
		m.onFocusChanged(),
		m.displayNotificationMsg(fmt.Sprintf("saved \"%s\"", updatedHost.Title())),
	)
}

func (m *listModel) setItemAndReorder(newIndex, currentIndex int, host ListItemHost) tea.Cmd {
	m.Model.RemoveItem(currentIndex)
	cmd := m.Model.InsertItem(newIndex, host)

	// The collection is not yet updated and m.VisibleItems() may NOT contain the updated host yet,
	// filtering is enabled. However, we must predict the new index.
	visibleTitles := lo.Reduce(m.VisibleItems(), func(agg []string, item list.Item, index int) []string {
		i := item.(ListItemHost)
		return lo.Ternary(i.ID == host.ID, agg, append(agg, i.uniqueName()))
	}, []string{host.uniqueName()})

	slices.Sort(visibleTitles)
	newVisibleItemsIndex := slices.Index(visibleTitles, host.uniqueName())

	m.Select(newVisibleItemsIndex)

	return cmd
}

func (m *listModel) onHostCreated(msg message.HostCreated) tea.Cmd {
	createdHostItem := ListItemHost{Host: msg.Host}
	titles := lo.Reduce(m.Items(), func(agg []string, item list.Item, index int) []string {
		return append(agg, item.(ListItemHost).uniqueName())
	}, []string{createdHostItem.uniqueName()})

	slices.Sort(titles)
	index := lo.IndexOf(titles, createdHostItem.uniqueName())
	cmd := m.Model.InsertItem(index, createdHostItem)

	// ResetFilter is required here because, user can create a new Item which will be filtered out,
	// therefore the user will not see any changes in the UI which is confusing.
	m.ResetFilter()
	// m.Select requires a visible item index, but because we reset filter VisibleItems array equals to Items
	m.Select(index)

	return tea.Sequence(
		// If host position coincides with other host, then let the underlying model to handle that
		cmd,
		m.onFocusChanged(),
		m.displayNotificationMsg(fmt.Sprintf("created \"%s\"", createdHostItem.Title())),
	)
}

func (m *listModel) onFocusChanged() tea.Cmd {
	m.updateTitle()
	m.updateKeyMap()

	if hostItem, ok := m.SelectedItem().(ListItemHost); ok {
		m.logger.Debug("[UI] Focus changed to host id: %v, title: %s", hostItem.ID, hostItem.Title())

		return tea.Sequence(
			message.TeaCmd(message.HostSelected{HostID: hostItem.ID}),
			message.TeaCmd(message.RunProcessSSHLoadConfig{Host: hostItem.Host}),
		)
	}

	m.logger.Debug("[UI] Focus is not set to any item in the list")
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

func (m *listModel) onToggleLayout() tea.Cmd {
	m.updateChildModel(msgToggleLayout{m.appState.ScreenLayout})
	// When switch between screen layouts, it's required to update pagination.
	// ListModel's updatePagination method is private and cannot be called from
	// here. One of the ways to trigger it is to invoke model.SetSize method.
	m.Model.SetSize(m.Width(), m.Height())
	// Another hack - need to invoke inner model's private updateKeyBindings method
	// when screen layout is toggled. If not update keybindings, then pagination
	// keys stop working properly when switching between compact and normal layouts.
	// The only way to trigger updateKeyBindings is via SetFilteringEnabled method.
	// The alternative way is copy and paste the entire updateKeyBindings method into
	// this file, and invoke it from here.
	m.SetFilteringEnabled(m.FilteringEnabled())
	notificationMsg := map[constant.ScreenLayout]string{
		constant.ScreenLayoutDescription: "show description",
		constant.ScreenLayoutGroup:       "group view",
		constant.ScreenLayoutCompact:     "compact view",
	}[m.appState.ScreenLayout]

	return m.displayNotificationMsg(notificationMsg)
}

/*
 * Helper methods.
 */

func (m *listModel) constructProcessCmd(processType constant.ProcessType) tea.Cmd {
	// Do not use m.SelectedItem() here!
	// list.Model keeps 2 collections - m.items and m.filteredItems, which can be inconsistent
	// as a result in some hosts taken from m.filteredItems ssh config is nil.
	var host *hostModel.Host
	for _, item := range m.Items() {
		if listItemHost, ok := item.(ListItemHost); ok && listItemHost.ID == m.appState.Selected {
			host = &listItemHost.Host
			break
		}
	}

	if host == nil {
		m.logger.Error("[UI] Could not find host with ID='%d'", m.appState.Selected)
		return message.TeaCmd(message.ErrorOccurred{Err: errors.New(itemNotSelectedErrMsg)})
	}

	if host.SSHClientConfig == nil {
		errorText := fmt.Sprintf("[UI] SSH config is not set for host ID='%d', Title='%s'", host.ID, host.Title)
		m.logger.Error(errorText)
		return message.TeaCmd(message.ErrorOccurred{Err: errors.New(errorText)})
	}

	if processType == constant.ProcessTypeSSHConnect {
		return message.TeaCmd(message.RunProcessSSHConnect{Host: *host})
	} else if processType == constant.ProcessTypeSSHCopyID {
		return message.TeaCmd(message.RunProcessSSHCopyID{Host: *host})
	}

	return nil
}

func (m *listModel) updateTitle() {
	var newTitle string
	item, isHost := m.SelectedItem().(ListItemHost)

	switch {
	case m.mode == modeSSHCopyID && isHost:
		newTitle = m.Styles.Title.Render("copy ssh key to the remote host? (y/N)")
	case m.mode == modeRemoveItem && isHost:
		newTitle = fmt.Sprintf("delete \"%s\"? (y/N)", item.Title())
		newTitle = m.Styles.Title.Render(newTitle)
	case m.mode == modeCloseApp:
		newTitle = m.Styles.Title.Render("close app? (y/N)")
	case isHost:
		// Replace Windows ssh prefix "cmd /c ssh" with "ssh"
		connectCmd := strings.Replace(item.Host.CmdSSHConnect(), "cmd /c ", "", 1)
		newTitle = m.prefixWithGroupName(connectCmd)
	default:
		// If it's NOT a host list item, then probably the list is just empty
		newTitle = m.prefixWithGroupName(defaultListTitle)
	}

	if m.Title != newTitle {
		m.Title = newTitle
		m.logger.Debug("[UI] New list title: %s", newTitle)
	}
}

func (m *listModel) prefixWithGroupName(title string) string {
	if !utils.StringEmpty(&m.appState.Group) {
		shortGroupName := utils.StringAbbreviation(m.appState.Group)
		shortGroupName = m.Styles.Group.Render(shortGroupName)
		return fmt.Sprintf("%s%s", shortGroupName, m.Styles.Title.Render(title))
	}

	return m.Styles.Title.Render(title)
}

func (m *listModel) updateKeyMap() {
	shouldShowEditButtons := m.SelectedItem() != nil

	if shouldShowEditButtons != m.keyMap.ShouldShowEditButtons() {
		m.logger.Debug("[UI] Show edit keyboard shortcuts: %v", shouldShowEditButtons)
		m.keyMap.SetShouldShowEditButtons(shouldShowEditButtons)
	}
}

func (m *listModel) selectHostByID(id int) tea.Cmd {
	// Use VisibleItems() instead of Items() because we need to find the correct index when deleting an item
	// while in filter mode where part of the collection is hidden. You can replicate a wrong behavior when using Items():
	// Enter filter mode, enter remove mode and then cancel it. The focus will be lost.
	_, index, found := lo.FindIndexOf(m.VisibleItems(), func(item list.Item) bool {
		hostItem, ok := item.(ListItemHost)
		return ok && hostItem.ID == id
	})

	if found {
		// Here we should check if the item with this index is already selected.
		// However, this will cause problems with title update when we enter remove
		// mode and then cancel it.
		m.Select(index)
	} else {
		// If host is not found, then we should reset the focus. Function 'onFocusChanged', will update application state.
		m.Select(0)
	}

	// This is a side effect. Ideally, it should not be here.
	return m.onFocusChanged()
}

/*
 * Deal with actions which require confirmation from the user.
 */

func (m *listModel) enterSSHCopyIDMode() tea.Cmd {
	// Check if item is selected.
	_, ok := m.SelectedItem().(ListItemHost)
	if !ok {
		m.logger.Debug("[UI] Cannot copy id. Host is not selected.")
		return message.TeaCmd(message.ErrorOccurred{Err: errors.New(itemNotSelectedErrMsg)})
	}

	m.mode = modeSSHCopyID
	m.logger.Debug("[UI] Enter %s mode. Ask user for confirmation.", m.mode)
	m.updateTitle()

	return nil
}

func (m *listModel) enterRemoveItemMode() tea.Cmd {
	// Check if item is selected.
	_, ok := m.SelectedItem().(ListItemHost)
	if !ok {
		m.logger.Debug("[UI] Cannot remove. Host is not selected.")
		return message.TeaCmd(message.ErrorOccurred{Err: errors.New(itemNotSelectedErrMsg)})
	}

	m.mode = modeRemoveItem
	m.logger.Debug("[UI] Enter %s mode. Ask user for confirmation.", m.mode)
	m.updateTitle()

	return nil
}

func (m *listModel) enterCloseAppMode() {
	m.mode = modeCloseApp
	m.logger.Debug("[UI] Enter %s mode. Ask user for confirmation.", m.mode)
	m.updateTitle()
}

func (m *listModel) handleKeyEventWhenModeEnabled(msg tea.KeyMsg) tea.Cmd {
	if key.Matches(msg, m.keyMap.confirm) {
		return m.confirmAction()
	}

	// If user doesn't confirm the operation, we go back to normal mode and update
	// title back to normal, this exact key event won't be handled
	m.logger.Debug("[UI] Exit %s mode. Cancel action.", m.mode)

	if hostListItem, ok := m.SelectedItem().(ListItemHost); ok {
		m.mode = modeDefault
		return m.selectHostByID(hostListItem.ID)
	}

	m.logger.Error("[UI] Exit %s mode, but cannot set focus on an item in the list of hosts.", m.mode)
	m.mode = modeDefault
	m.updateTitle()
	return nil
}

func (m *listModel) confirmAction() tea.Cmd {
	m.logger.Debug("[UI] Exit %s mode. Confirm action.", m.mode)

	var cmd tea.Cmd
	if m.mode == modeRemoveItem { //nolint:gocritic // better readable without switch
		m.mode = modeDefault
		cmd = m.removeItem() // removeItem triggers title and keymap updates. See "onFocusChanged" method.
	} else if m.mode == modeSSHCopyID {
		m.mode = modeDefault
		m.updateTitle()
		cmd = m.constructProcessCmd(constant.ProcessTypeSSHCopyID)
	} else if m.mode == modeCloseApp {
		m.mode = modeDefault
		cmd = tea.Quit
	}

	return cmd
}

func (m *listModel) displayNotificationMsg(msg string) tea.Cmd {
	if utils.StringEmpty(&msg) {
		return nil
	}

	m.logger.Debug("[UI] Notification message: %s", msg)
	m.Title = m.Styles.Title.Render(msg)
	if m.notificationMessageTimer != nil {
		m.notificationMessageTimer.Stop()
	}

	m.notificationMessageTimer = time.NewTimer(m.notificationMessageTimeout)

	return func() tea.Msg {
		<-m.notificationMessageTimer.C
		return msgHideNotification{}
	}
}
