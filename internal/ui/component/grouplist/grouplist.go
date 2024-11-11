package grouplist

import (
	"context"
	"slices"

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

type (
	msgErrorOccurred struct{ err error }
)

type ListModel struct {
	list.Model
	repo      storage.HostStorage
	appState  *state.ApplicationState
	logger    iLogger
	groupList []string
}

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
	m.logger.Debug("[UI] Load groups from the database")
	hosts, err := m.repo.GetAll()
	if err != nil {
		m.logger.Error("[UI] Cannot read database. %v", err)
		return message.TeaCmd(msgErrorOccurred{err}) // TODO: msgErrorOccurred should be public and shared between hostlist and grouplist ?
	}

	groupList := lo.Map(hosts, func(h host.Host, index int) string {
		return h.Group
	})

	slices.Sort(groupList)
	// m.groupList = groupList
	m.groupList = []string{"group 1", "group 2"}

	return nil
}

func (m *ListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.Model, cmd = m.Model.Update(msg)

	return m, cmd
}

func (m *ListModel) View() string {
	return m.Model.View()
}
