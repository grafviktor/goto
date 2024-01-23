// Package hostlist implements the host list view.
package hostlist

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
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
	"github.com/grafviktor/goto/internal/utils"
	"github.com/grafviktor/goto/internal/utils/ssh"
)

var (
	docStyle               = lipgloss.NewStyle().Margin(1, 2)
	itemNotSelectedMessage = "you must select an item"
	modeRemoveItem         = "removeItem"
)

type logger interface {
	Debug(format string, args ...any)
	Error(format string, args ...any)
}

type (
	// MsgEditItem fires when user press edit button.
	MsgEditItem struct{ HostID int }
	// MsgCopyItem fires when user press copy button.
	MsgCopyItem struct{ HostID int }
	// MsgNewItem fires when user press new host button.
	MsgNewItem      struct{}
	msgInitComplete struct{}
	msgErrorOccured struct{ err error }
	// MsgRepoUpdated - fires when data layer updated and it's required to reload the host list.
	MsgRepoUpdated struct{}
	msgRefreshUI   struct{}
)

type listModel struct {
	innerModel list.Model
	repo       storage.HostStorage
	keyMap     *keyMap
	appState   *state.ApplicationState
	logger     logger
	mode       string
}

// New - creates new host list model.
// context - is not used.
// storage - is the data layer.
// appState - is the application state, usually we want to restore previous state when application restarts,
// for instance focus previously selected host.
// log - application logger.
func New(_ context.Context, storage storage.HostStorage, appState *state.ApplicationState, log logger) listModel {
	delegate := list.NewDefaultDelegate()
	delegateKeys := newDelegateKeyMap()
	listItems := []list.Item{}
	m := listModel{
		innerModel: list.New(listItems, delegate, 0, 0),
		keyMap:     delegateKeys,
		repo:       storage,
		appState:   appState,
		logger:     log,
	}

	m.innerModel.KeyMap.CursorUp.Unbind()
	m.innerModel.KeyMap.CursorUp = delegateKeys.cursorUp
	m.innerModel.KeyMap.CursorDown.Unbind()
	m.innerModel.KeyMap.CursorDown = delegateKeys.cursorDown

	m.innerModel.AdditionalShortHelpKeys = delegateKeys.ShortHelp
	m.innerModel.AdditionalFullHelpKeys = delegateKeys.FullHelp

	m.innerModel.Title = "press 'n' to add a new host"
	m.innerModel.SetShowStatusBar(false)

	return m
}

func (m listModel) Init() tea.Cmd {
	return tea.Batch(message.TeaCmd(msgInitComplete{}))
}

func (m listModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.innerModel.SettingFilter() {
			// If filter is enabled, we should not handle any keyboard messages,
			// as it should be done by filter component.

			// However, there is one special case, which should be taken into account:
			// When user filter's out values and presses down key on her keyboard
			// we need to ensure that the title contains proper selection.
			// that's why we need to invoke title update function.
			// See https://github.com/grafviktor/goto/issues/37
			m = m.listTitleUpdate()
			break
		}

		if m.mode != "" {
			// Handle key event when some mode is enabled. For instance "removeMode".
			return m.handleKeyEventWhenModeEnabled(msg)
		}

		switch {
		case key.Matches(msg, m.keyMap.connect):
			return m.executeCmd(msg)
		case key.Matches(msg, m.keyMap.remove):
			return m.enterRemoveItemMode()
		case key.Matches(msg, m.keyMap.edit):
			return m.editItem(msg)
		case key.Matches(msg, m.keyMap.append):
			return m, message.TeaCmd(MsgEditItem{})
		case key.Matches(msg, m.keyMap.clone):
			return m.copyItem(msg)
		}

		// Dispatch msgRefreshUI message to update list title.
		// Actually we only need to dispatch it when we switch between list items
		cmds = append(cmds, message.TeaCmd(msgRefreshUI{}))
	case tea.WindowSizeMsg:
		// triggers immediately after app start because we render this component by default
		h, v := docStyle.GetFrameSize()
		m.innerModel.SetSize(msg.Width-h, msg.Height-v)
		m.logger.Debug("[UI] Set host list size: %d %d", m.innerModel.Width(), m.innerModel.Height())
	case MsgRepoUpdated:
		m.logger.Debug("[UI] Load hostnames from the database")
		return m.refreshRepo(msg)
	case msgInitComplete:
		m.logger.Debug("[UI] Load hostnames from the database")
		return m.refreshRepo(msg)
	case msgRefreshUI:
		m = m.listTitleUpdate()
		m.logger.Debug("[UI] Update hostlist title: %s", m.innerModel.Title)
		var cmd tea.Cmd
		m, cmd = m.onFocusChanged(msg)
		cmds = append(cmds, cmd)
		return m, tea.Batch(cmds...)
	}

	// If we could not find our own update handler, we pass message to the original model
	// otherwise we would have to implement all key hanlders and other stuff by ourselves
	var innerModelCmd tea.Cmd
	m.innerModel, innerModelCmd = m.innerModel.Update(msg)
	cmds = append(cmds, innerModelCmd)
	return m, tea.Batch(cmds...)
}

func (m listModel) View() string {
	return docStyle.Render(m.innerModel.View())
}

func (m listModel) handleKeyEventWhenModeEnabled(msg tea.KeyMsg) (listModel, tea.Cmd) {
	if key.Matches(msg, m.keyMap.confirm) {
		m.logger.Debug("[UI] User confirms action")
		return m.confirmAction()
	}

	// If user doesn't confirm the operation, we go back to normal mode and update
	// title back to normal, this exact key event won't be handled
	m.logger.Debug("[UI] User cancels action")
	m.mode = ""
	return m.listTitleUpdate(), nil
}

func (m listModel) confirmAction() (listModel, tea.Cmd) {
	if m.mode == modeRemoveItem {
		m.logger.Debug("[UI] Exit remove mode")
		m.mode = ""
		return m.removeItem()
	}

	return m, nil
}

func (m listModel) enterRemoveItemMode() (listModel, tea.Cmd) {
	// Check if item is selected.
	item, ok := m.innerModel.SelectedItem().(ListItemHost)
	if !ok {
		return m, message.TeaCmd(msgErrorOccured{err: errors.New(itemNotSelectedMessage)})
	}

	host := item.Unwrap()
	m.logger.Debug("[UI] Enter remove mode. Selected host id: %d, title %s", host.ID, host.Title)
	m.mode = modeRemoveItem

	return m, message.TeaCmd(msgRefreshUI{})
}

func (m listModel) removeItem() (listModel, tea.Cmd) {
	m.logger.Debug("[UI] Remove host from the database")
	item, ok := m.innerModel.SelectedItem().(ListItemHost)
	if !ok {
		m.logger.Error("[UI] Cannot cast selected item to host model")
		return m, message.TeaCmd(msgErrorOccured{err: errors.New(itemNotSelectedMessage)})
	}

	err := m.repo.Delete(item.ID)
	if err != nil {
		m.logger.Debug("[UI] Error removing host from the database. %v", err)
		return m, message.TeaCmd(msgErrorOccured{err})
	}

	return m, tea.Batch(
		message.TeaCmd(MsgRepoUpdated{}),
		message.TeaCmd(msgRefreshUI{}),
	)
}

func (m listModel) refreshRepo(_ tea.Msg) (listModel, tea.Cmd) {
	items := []list.Item{}
	hosts, err := m.repo.GetAll()
	if err != nil {
		m.logger.Error("[UI] Cannot read database. %v", err)
		return m, message.TeaCmd(msgErrorOccured{err})
	}

	slices.SortFunc(hosts, func(a, b model.Host) int {
		if a.Title < b.Title {
			return -1
		}
		return 1
	})

	// Wrap hosts into List items
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

	return m, tea.Batch(setItemsCmd, message.TeaCmd(msgRefreshUI{}))
}

func (m listModel) editItem(_ tea.Msg) (listModel, tea.Cmd) {
	item, ok := m.innerModel.SelectedItem().(ListItemHost)
	if !ok {
		return m, message.TeaCmd(msgErrorOccured{err: errors.New(itemNotSelectedMessage)})
	}

	host := *item.Unwrap()
	m.logger.Debug("[UI] Edit item id: %d, title: %s", host.ID, host.Title)
	return m, message.TeaCmd(MsgEditItem{HostID: host.ID})
}

func (m listModel) copyItem(_ tea.Msg) (listModel, tea.Cmd) {
	item, ok := m.innerModel.SelectedItem().(ListItemHost)
	if !ok {
		m.logger.Error("[UI] Cannot cast selected item to host model")
		return m, message.TeaCmd(msgErrorOccured{err: errors.New(itemNotSelectedMessage)})
	}

	originalHost := item.Unwrap()
	m.logger.Debug("[UI] Duplicate selected item id: %d, %s", originalHost.ID, originalHost.Title)
	clonedHost := originalHost.Clone()
	for i := 1; ok; i++ {
		clonedHostTitle := fmt.Sprintf("%s %d", originalHost.Title, i)
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
		return m, message.TeaCmd(msgErrorOccured{err})
	}

	return m, tea.Batch(
		message.TeaCmd(MsgRepoUpdated{}),
		message.TeaCmd(msgRefreshUI{}),
	)
}

func (m listModel) buildProcess(errorWriter *stdErrorWriter) (*exec.Cmd, error) {
	m.logger.Debug("[UI] Build external command")
	item, ok := m.innerModel.SelectedItem().(ListItemHost)
	if !ok {
		return nil, errors.New(itemNotSelectedMessage)
	}

	host := *item.Unwrap()
	command := ssh.ConstructCMD(ssh.BaseCMD(), utils.HostModelToOptionsAdaptor(host)...)
	process := utils.BuildProcess(command)
	process.Stdout = os.Stdout
	process.Stderr = errorWriter

	return process, nil
}

func (m listModel) runProcess(process *exec.Cmd, errorWriter *stdErrorWriter) (listModel, tea.Cmd) {
	m.logger.Debug("[UI] Prepare external process")
	execCmd := tea.ExecProcess(process, func(err error) tea.Msg {
		// This callback triggers when external process exits
		if err != nil {
			m.logger.Error("[UI] External process build error. %v", err)
			errorMessage := strings.TrimSpace(string(errorWriter.err))
			if utils.StringEmpty(errorMessage) {
				errorMessage = err.Error()
			}

			commandWhichFailed := strings.Join(process.Args, " ")
			// errorDetails contains command which was executed and the error text.
			errorDetails := fmt.Sprintf("Command: %s\nError:   %s", commandWhichFailed, errorMessage)
			return message.RunProcessErrorOccured{Err: errors.New(errorDetails)}
		}

		return nil
	})

	return m, execCmd
}

func (m listModel) executeCmd(_ tea.Msg) (listModel, tea.Cmd) {
	m.logger.Debug("[UI] Run external process")
	errorWriter := stdErrorWriter{}
	process, err := m.buildProcess(&errorWriter)
	if err != nil {
		m.logger.Debug("[UI] Error running external process. %v", err)
		return m, message.TeaCmd(msgErrorOccured{err: errors.New(itemNotSelectedMessage)})
	}

	return m.runProcess(process, &errorWriter)
}

func (m listModel) listTitleUpdate() listModel {
	item, ok := m.innerModel.SelectedItem().(ListItemHost)
	if !ok {
		return m
	}

	if m.mode == modeRemoveItem {
		m.innerModel.Title = fmt.Sprintf("delete \"%s\" ? (y/N)", item.Title())
		return m
	}

	m.innerModel.Title = ssh.ConstructCMD("ssh", utils.HostModelToOptionsAdaptor(*item.Unwrap())...)
	return m
}

func (m listModel) onFocusChanged(_ tea.Msg) (listModel, tea.Cmd) {
	if hostItem, ok := m.innerModel.SelectedItem().(ListItemHost); ok {
		return m, message.TeaCmd(message.HostListSelectItem{HostID: hostItem.ID})
	}

	return m, nil
}

// stdErrorWriter - is an object which pretends to be a writer, however it saves all data into 'err' variable
// for future reading and do not write anything in terminal. We need it to display a formatted error in the console
// when it's required, but not when it's done by default.
type stdErrorWriter struct {
	err []byte
}

// Write - doesn't write anything, it saves all data in err variable, which can ve read later.
func (writer *stdErrorWriter) Write(p []byte) (n int, err error) {
	writer.err = append(writer.err, p...)

	// Hide error from the console, otherwise it will be seen in a subsequent ssh calls
	// To return to default behavior use: return os.Stderr.Write(p)
	// We must return the number of bytes which were written using `len(p)`,
	// otherwise exec.go will throw 'short write' error.
	return len(p), nil
}
