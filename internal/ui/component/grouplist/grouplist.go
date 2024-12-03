package grouplist

import (
	"context"
	"slices"

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

type (
	msgErrorOccurred struct{ err error }
)

type ListModel struct {
	list.Model
	repo      storage.HostStorage
	appState  *state.ApplicationState
	logger    iLogger
}

var docStyle = lipgloss.NewStyle().Margin(1, 2)

func New(_ context.Context, storage storage.HostStorage, appState *state.ApplicationState, log iLogger) *ListModel {
	var listItems []list.Item
	model := list.New(listItems, list.NewDefaultDelegate(), 0, 0)
	// This line affects sorting when filtering enabled. What UnsortedFilter
	// does - it filters the collection, but leaves initial items order unchanged.
	// Default filter on the contrary - filters the collection based on the match rank.
	model.Filter = list.UnsortedFilter

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
		// FIXME: Never gets called!
		h, v := docStyle.GetFrameSize()
		m.SetSize(msg.Width-h, msg.Height-v)
		m.logger.Debug("[UI] Set host list size: %d %d", m.Width(), m.Height())
		return m, nil
	case tea.KeyMsg:
		cmd = m.handleKeyboardEvent(msg)
		return m, cmd
	case 	message.OpenSelectGroupForm:
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
		return message.TeaCmd(message.CloseSelectGroupForm{})
	case tea.KeyEnter:
		return message.TeaCmd(message.CloseSelectGroupForm{})
	}

	m.Model, cmd = m.Model.Update(msg)
	return cmd
}

func (m *ListModel) loadHostGroups() tea.Cmd {
	m.logger.Debug("[UI] Load groups from the database")
	hosts, err := m.repo.GetAll()
	if err != nil {
		m.logger.Error("[UI] Cannot read database. %v", err)
		return message.TeaCmd(msgErrorOccurred{err}) // TODO: msgErrorOccurred should be public and shared between hostlist and grouplist ?
	}

	groupList := []string{}
	groupList = append(groupList, "default")
	groupList = append(groupList, "default1")
	groupList = append(groupList, "default2")
	groupList = append(groupList, "default3")
	groupList = append(groupList, "default4")

	lo.Map(hosts, func(h host.Host, index int) string {
		return h.Group
	})
	lo.Uniq(groupList)
	slices.Sort(groupList)
	// m.groupList = groupList
	// m.groupList = []string{"group 1", "group 2"}

	// m.SetItems([]ListItemHostGroup{"group1", "group2"})

	items := make([]list.Item, 0, len(groupList))
	for _, group := range groupList {
		items = append(items, ListItemHostGroup{group})
	}

	return m.SetItems(items)
}