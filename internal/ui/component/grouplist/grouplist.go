// Package grouplist implements the group list view. Which user can use to select active group.
// Based on a selected group different set of hosts will be shown.
package grouplist

import (
	"context"
	"slices"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
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

var noGroupSelected = "~ all ~"

type Model struct {
	list.Model

	repo     storage.HostStorage
	appState *state.State
	logger   iLogger
	styles   styles
}

// New - creates a new UI component which is used to select a host group from a list,
// with pre-defined initial parameters.
func New(_ context.Context, repo storage.HostStorage, appState *state.State, log iLogger) *Model {
	styles := defaultStyles()

	var listItems []list.Item
	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = false
	delegate.Styles = styles.styleListDelegate
	delegate.SetSpacing(0)

	model := list.New(listItems, delegate, 0, 0)
	model.DisableQuitKeybindings() // We don't want to quit the app from this view.

	// Setup model styles.
	model.Styles = styles.styleList
	model.FilterInput.PromptStyle = styles.stylePrompt
	model.FilterInput.TextStyle = styles.styleFilterInput
	model.Paginator.ActiveDot = styles.stylePaginatorActiveDot
	model.Paginator.InactiveDot = styles.stylePaginatorInactiveDot
	model.Help.Styles = styles.styleHelp

	m := Model{
		Model:    model,
		repo:     repo,
		appState: appState,
		logger:   log,
		styles:   styles,
	}

	m.Title = "select group"

	return &m
}

func (m *Model) Init() tea.Cmd { return nil }

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h, v := m.styles.styleComponentMargins.GetFrameSize()
		m.SetSize(msg.Width-h, msg.Height-v)
		m.logger.Debug("[UI] Set group list size: %d %d", m.Width(), m.Height())
		return m, nil
	case tea.KeyMsg:
		cmd = m.handleKeyboardEvent(msg)
		cmds = append(cmds, cmd)
	case message.ViewGroupListOpen:
		return m, m.loadItems()
	}

	m.Model, cmd = m.Model.Update(msg)
	// Only calculate status bar visibility AFTER the model is updated.
	m.SetShowStatusBar(m.FilterState() != list.Unfiltered)

	return m, tea.Batch(append(cmds, cmd)...)
}

func (m *Model) View() string {
	return m.styles.styleComponentMargins.Render(m.Model.View())
}

func (m *Model) handleKeyboardEvent(msg tea.KeyMsg) tea.Cmd {
	//exhaustive:ignore // Handle only specific keys
	switch msg.Type {
	case tea.KeyEscape:
		return m.handleEscapeKey()
	case tea.KeyEnter:
		return m.handleEnterKey()
	}

	return nil
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

	m.logger.Debug("[UI] Load complete. Found '%d' groups", len(groupList))
	slices.Sort(groupList)
	// noGroupSelected always comes first
	groupList = append([]string{noGroupSelected}, groupList...)

	items := make([]list.Item, 0, len(groupList))
	for _, group := range groupList {
		items = append(items, ListItemHostGroup{group})
	}

	return m.SetItems(items)
}

func (m *Model) handleEscapeKey() tea.Cmd {
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

func (m *Model) handleEnterKey() tea.Cmd {
	// If filter is active, by default pressing Enter just selects
	// the first item from the list of filtered items. Prevent that.
	if m.FilterState() == list.Filtering {
		m.logger.Debug("[UI] Enter key. Select item in group list view.")
		return nil
	}

	// Otherwise, select only visible item, going to hostlist view.
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
