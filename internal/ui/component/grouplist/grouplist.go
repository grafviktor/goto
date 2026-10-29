// Package grouplist implements the group list view. Which user can use to select active group.
// Based on a selected group different set of hosts will be shown.
package grouplist

import (
	"context"
	"slices"
	"strings"
	"time"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/samber/lo"

	"github.com/grafviktor/goto/internal/model/host"
	"github.com/grafviktor/goto/internal/state"
	"github.com/grafviktor/goto/internal/storage"
	"github.com/grafviktor/goto/internal/ui/message"
)

type iLogger interface {
	Debug(format string, args ...any)
	Info(format string, args ...any)
	Error(format string, args ...any)
}

const (
	noGroupSelected          = "~ all ~"
	timeToWaitForSecondClick = 500 * time.Millisecond
)

type msgSecondClickTimeOut struct{}

type Model struct {
	list.Model

	repo             storage.HostStorage
	appState         *state.State
	logger           iLogger
	styles           styles
	additionalKeyMap *keyMap
	// For handling double click
	waitForSecondClick bool
}

// New - creates a new UI component which is used to select a host group from a list,
// with pre-defined initial parameters.
func New(_ context.Context, repo storage.HostStorage, appState *state.State, log iLogger) *Model {
	styles := defaultStyles()

	var listItems []list.Item
	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = false
	delegate.Styles = styles.listDelegate
	delegate.SetSpacing(0)

	model := list.New(listItems, delegate, 0, 0)
	model.DisableQuitKeybindings() // We don't want to quit the app from this view.

	// Setup filter input styles.
	filterStyles := model.FilterInput.Styles()
	filterStyles.Focused.Prompt = styles.prompt
	filterStyles.Focused.Text = styles.filterInput
	model.FilterInput.SetStyles(filterStyles)

	// Setup model styles.
	model.Styles = styles.list
	model.Paginator.ActiveDot = styles.paginatorActiveDot
	model.Paginator.InactiveDot = styles.paginatorInactiveDot
	model.Help.Styles = styles.help

	m := Model{
		Model:    model,
		repo:     repo,
		appState: appState,
		logger:   log,
		styles:   styles,
	}

	m.Title = "select group"

	m.additionalKeyMap = newKeyMap()
	m.AdditionalShortHelpKeys = m.additionalKeyMap.ShortHelp
	m.AdditionalFullHelpKeys = m.additionalKeyMap.FullHelp

	return &m
}

func (m *Model) Init() tea.Cmd {
	return m.loadItems()
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h, v := m.styles.componentMargins.GetFrameSize()
		m.SetSize(msg.Width-h, msg.Height-v)
		m.logger.Debug("[UI] Set group list size: %d %d", m.Width(), m.Height())
		return m, nil
	case tea.KeyPressMsg:
		cmd = m.handleKeyboardEvent(msg)
		cmds = append(cmds, cmd)
	case tea.MouseMsg:
		return m, m.handleMouseMsg(msg)
	case msgSecondClickTimeOut:
		m.waitForSecondClick = false
		return m, nil
	case message.ViewGroupListOpen:
		return m, m.loadItems()
	}

	m.Model, cmd = m.Model.Update(msg)
	// Only calculate status bar visibility AFTER the model is updated.
	m.SetShowStatusBar(m.FilterState() != list.Unfiltered)

	return m, tea.Batch(append(cmds, cmd)...)
}

func (m *Model) View() tea.View {
	return tea.NewView(m.styles.componentMargins.Render(m.Model.View()))
}

func (m *Model) handleKeyboardEvent(msg tea.KeyPressMsg) tea.Cmd {
	//exhaustive:ignore // Handle only specific keys, other events are handled by the list model.
	switch {
	case key.Matches(msg, m.additionalKeyMap.selectGroup):
		return m.handleSelectGroup()
	case key.Matches(msg, m.additionalKeyMap.deselectGroup):
		return m.handleDeselectGroup()
	case key.Matches(msg, m.additionalKeyMap.quitApp):
		if m.FilterState() != list.Filtering {
			return tea.Quit
		}
	}

	return nil
}

func (m *Model) handleSelectGroup() tea.Cmd {
	// If filter is active, by default pressing Enter just selects
	// the first item from the list of filtered items. Prevent that.
	if m.FilterState() == list.Filtering {
		m.logger.Debug("[UI] Enter key. Select item in group list view.")
		return nil
	}

	if m.SelectedItem() == nil {
		m.logger.Debug("[UI] Enter key. No group selected, nothing to do.")
		return nil
	}

	return m.selectGroupAndCloseView()
}

func (m *Model) selectGroupAndCloseView() tea.Cmd {
	selected := m.SelectedItem().(ListItemHostGroup).Title() //nolint:errcheck // SelectedItem always returns ListItemHostGroup
	selected = strings.TrimSpace(selected)

	if selected == noGroupSelected {
		selected = ""
	}

	m.logger.Debug("[UI] Enter key. Select group '%s' and exit from group list view.", selected)
	return tea.Sequence(
		message.TeaCmd(message.GroupSelect{Name: selected}),
		message.TeaCmd(message.ViewGroupListClose{}),
	)
}

func (m *Model) handleDeselectGroup() tea.Cmd {
	// If model is in filter mode and press ESC, just disable filtering.
	if m.FilterState() == list.Filtering || m.FilterState() == list.FilterApplied {
		m.logger.Debug("[UI] Escape key. Deactivate filter in group list view.")
		return nil
	}

	// If group view is shown and user presses ESC, we should
	// deselect the group view and then show the full host list.
	m.logger.Debug("[UI] Escape key. Deselect group and exit from group list view.")
	return tea.Sequence(
		message.TeaCmd(message.GroupSelect{Name: ""}),
		message.TeaCmd(message.ViewGroupListClose{}),
	)
}

func (m *Model) handleMouseMsg(msg tea.MouseMsg) tea.Cmd {
	if m.FilterState() == list.Filtering {
		return nil
	}

	if msg, ok := msg.(tea.MouseClickMsg); ok {
		if msg.Mouse().Button == tea.MouseLeft {
			return m.focusItem(msg)
		}
	}

	return nil
}

func (m *Model) focusItem(msg tea.MouseClickMsg) tea.Cmd {
	if m.paginatorRowPosition() == msg.Mouse().Y {
		m.handlePaginatorClick()
		return nil
	}

	index := m.itemAt(msg.Mouse().Y)
	if index < 0 {
		return nil
	}

	if index == m.Index() && m.waitForSecondClick {
		m.waitForSecondClick = false
		return m.selectGroupAndCloseView()
	}

	m.Select(index)
	m.waitForSecondClick = true
	return tea.Tick(timeToWaitForSecondClick, func(_ time.Time) tea.Msg {
		return msgSecondClickTimeOut{}
	})
}

func (m *Model) paginatorRowPosition() int {
	if m.Paginator.TotalPages <= 1 {
		return -1
	}

	paginationStyle := m.Styles.PaginationStyle
	// This tweak is taken from charm.land/bubbles/v2@v2.1.0/list/list.go lines 1195-1200
	if paginationStyle.GetMarginTop() == 0 {
		paginationStyle = paginationStyle.MarginTop(1)
	}

	paginationHeight := lipgloss.Height(paginationStyle.Render(m.Paginator.View()))
	helpHeight := lipgloss.Height(m.Styles.HelpStyle.Render(m.Help.View(m.Model)))
	// Take the full height plus top margin and subtract help and pagination heights
	topSectionsHeight := m.styles.componentMargins.GetMarginTop() + m.Height() - helpHeight - paginationHeight
	// Height of the top sections plus paginator margin is becoming your paginator index.
	return topSectionsHeight + paginationStyle.GetMarginTop()
}

func (m *Model) handlePaginatorClick() {
	if m.Paginator.TotalPages > 1 && m.Paginator.OnLastPage() {
		for !m.Paginator.OnFirstPage() {
			m.Paginator.PrevPage()
		}
	} else {
		m.Paginator.NextPage()
	}
}

func (m *Model) itemAt(positionY int) int {
	// Find top panel height.
	listTop := m.styles.componentMargins.GetMarginTop()
	// There is no direct way to get height of the title, that's why we
	// simulate what render does and calculate its height manually.
	listTop += lipgloss.Height(m.Styles.TitleBar.Render(m.Styles.Title.Render(m.Title)))
	if m.ShowStatusBar() {
		// Need to render at least one letter to get real height.
		listTop += lipgloss.Height(m.Styles.StatusBar.Render("x"))
	}

	// terminalRow is just the position relative to the list start, excluding top sections.
	terminalRow := positionY - listTop
	if terminalRow < 0 {
		return -1
	}

	start, end := m.Paginator.GetSliceBounds(len(m.VisibleItems()))
	index := start + terminalRow
	if index >= end {
		return -1
	}
	return index
}

func (m *Model) loadItems() tea.Cmd {
	m.logger.Debug("[UI] Load groups from the database")
	hosts, err := m.repo.GetAll()
	if err != nil {
		m.logger.Error("[UI] Cannot read database. %v", err)
		return message.TeaCmd(message.ErrorOccurred{Err: err})
	}

	// Create a list of unique groups.
	groupList := []string{}
	lo.ForEach(hosts, func(h host.Host, _ int) {
		if strings.TrimSpace(h.Group) != "" {
			_, found := lo.Find(groupList, func(g string) bool {
				return strings.EqualFold(g, h.Group)
			})

			// Only add if there is no such group already in the list. Case ignored.
			if !found {
				groupList = append(groupList, h.Group)
			}
		}
	})

	groupCount := len(groupList)
	m.logger.Debug("[UI] Load complete. Found '%d' groups", groupCount)
	slices.Sort(groupList)
	// noGroupSelected always comes first
	groupList = append([]string{noGroupSelected}, groupList...)

	items := make([]list.Item, 0, len(groupList))
	for _, group := range groupList {
		items = append(items, ListItemHostGroup{group})
	}

	setItemsCmd := m.SetItems(items)
	m.selectSavedGroup()

	return tea.Sequence(
		setItemsCmd,
		message.TeaCmd(message.GroupListLoadComplete{GroupCount: groupCount}),
	)
}

// selectSavedGroup focuses the group stored in application state.
// An empty saved group, or one that is no longer present, stays on "~ all ~".
func (m *Model) selectSavedGroup() {
	saved := strings.TrimSpace(m.appState.Group)
	if saved == "" {
		m.Select(0)
		return
	}

	_, index, found := lo.FindIndexOf(m.Items(), func(item list.Item) bool {
		group, ok := item.(ListItemHostGroup)
		return ok && strings.EqualFold(group.Title(), saved)
	})

	if !found {
		m.Select(0)
		return
	}

	m.Select(index)
}
