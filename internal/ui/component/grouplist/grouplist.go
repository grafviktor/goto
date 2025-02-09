package grouplist

import (
	"context"
	"slices"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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

type ListModel struct {
	list.Model
	repo     storage.HostStorage
	appState *state.ApplicationState
	logger   iLogger
}

var docStyle = lipgloss.NewStyle().Margin(1, 2)
var unselectGroup = "* ALL *"

func New(_ context.Context, storage storage.HostStorage, appState *state.ApplicationState, log iLogger) *ListModel {
	var listItems []list.Item
	var listDelegate = list.NewDefaultDelegate()
	listDelegate.ShowDescription = false
	listDelegate.SetSpacing(0)
	model := list.New(listItems, listDelegate, 0, 0)

	m := ListModel{
		Model:    model,
		repo:     storage,
		appState: appState,
		logger:   log,
	}

	m.Title = "select group"
	m.SetShowStatusBar(false)

	return &m
}

func (m *ListModel) Init() tea.Cmd {
	return m.loadHostGroups()
}

func (m *ListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.SetSize(msg.Width-h, msg.Height-v)
		m.logger.Debug("[UI] Set group list size: %d %d", m.Width(), m.Height())
		return m, nil
	case tea.KeyMsg:
		cmd = m.handleKeyboardEvent(msg)
		return m, cmd
	case message.OpenSelectGroupForm:
		return m, m.loadHostGroups()
	}

	m.Model, cmd = m.Model.Update(msg)
	return m, cmd
}

func (m *ListModel) View() string {
	return docStyle.Render(m.Model.View())
}

func (m *ListModel) handleKeyboardEvent(msg tea.KeyMsg) tea.Cmd {
	var cmd tea.Cmd

	switch msg.Type {
	case tea.KeyEscape:
		return tea.Sequence(
			// If group view is shown and user presses ESC, we should
			// deselect the group view and then show the full host list.
			message.TeaCmd(message.GroupListSelectItem{GroupName: ""}),
			message.TeaCmd(message.CloseSelectGroupForm{}),
		)
	case tea.KeyEnter:
		selected := m.SelectedItem().(ListItemHostGroup).Title()
		selected = strings.TrimSpace(selected)

		if selected == unselectGroup {
			selected = ""
		}

		return tea.Batch(
			message.TeaCmd(message.GroupListSelectItem{GroupName: selected}),
			message.TeaCmd(message.CloseSelectGroupForm{}),
		)
	}

	m.Model, cmd = m.Model.Update(msg)
	return cmd
}

func (m *ListModel) loadHostGroups() tea.Cmd {
	m.logger.Debug("[UI] Load groups from the database")
	hosts, err := m.repo.GetAll()
	if err != nil {
		m.logger.Error("[UI] Cannot read database. %v", err)
		return message.TeaCmd(message.ErrorOccurred{Err: err})
	}

	// Create a list of unique groups, one group is always there - "unselectGroup".
	groupList := []string{unselectGroup}
	lo.ForEach(hosts, func(h host.Host, index int) {
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

	slices.Sort(groupList)

	items := make([]list.Item, 0, len(groupList))
	for _, group := range groupList {
		items = append(items, ListItemHostGroup{group})
	}

	return m.SetItems(items)
}
