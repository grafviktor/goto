package edithost

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/grafviktor/goto/internal/model"
	"github.com/grafviktor/goto/internal/state"
	"github.com/grafviktor/goto/internal/storage"
	"github.com/grafviktor/goto/internal/ui/component/hostlist"
	"github.com/grafviktor/goto/internal/ui/message"
	. "github.com/grafviktor/goto/internal/ui/message"
	"github.com/grafviktor/goto/internal/utils"
)

type Size struct {
	Width  int
	Height int
}

type MsgClose struct{}
type MsgSave struct{}

const ItemID string = "itemID"

// func New(ctx context.Context, storage storage.HostStorage, width int, height int) editModel {
func New(ctx context.Context, storage storage.HostStorage, state *state.ApplicationState) editModel {
	// if we can't cast host id to int, that means we're adding a new host. Ignoring the error
	hostID, _ := ctx.Value(ItemID).(int)
	host, err := storage.Get(hostID)
	if err != nil {
		host = model.Host{}
	}

	m := editModel{
		inputs:      make([]LabeledInput, 6),
		hostStorage: storage,
		host:        host,
		help:        help.New(),
		keyMap:      keys,
		appState:    state,
	}

	var t LabeledInput
	for i := range m.inputs {
		t = NewLabelInput()
		t.Cursor.Style = cursorStyle
		t.Placeholder = "n/a"

		switch i {
		case 0:
			t.Label = "Title"
			t.Focus()
			t.SetValue(host.Title)
		case 1:
			t.Label = "IP Address or Hostname"
			t.CharLimit = 128
			t.SetValue(host.Address)
		case 2:
			t.Label = "Description"
			t.CharLimit = 512
			t.SetValue(host.Description)
		case 3:
			t.Label = "Login"
			t.CharLimit = 128
			t.Placeholder = fmt.Sprintf("default: %s", utils.GetCurrentOSUser())
			t.SetValue(host.LoginName)
		case 4:
			t.Label = "Port"
			t.CharLimit = 5
			t.Placeholder = "default: 22"
			t.SetValue(host.RemotePort)
		case 5:
			t.Label = "Identity file path"
			t.CharLimit = 512
			t.Placeholder = "default: not set"
			t.SetValue(host.PrivateKeyPath)
		}

		m.inputs[i] = t
	}

	return m
}

type editModel struct {
	keyMap      keyMap
	hostStorage storage.HostStorage
	focusIndex  int
	inputs      []LabeledInput
	host        model.Host
	viewport    viewport.Model
	help        help.Model
	ready       bool
	appState    *state.ApplicationState
	// size        Size
}

func (m editModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m editModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	// forward keyboard events to help menu component
	m.help, cmd = m.help.Update(msg)
	cmds = append(cmds, cmd)

	// create or Update viewport
	m = m.updateViewPort(msg)

	switch msg := msg.(type) { //nolint:gocritic // we should use "switch" to receive message type
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keyMap.Save):
			m, cmd = m.save(msg)
			cmds = append(cmds, cmd)
		case key.Matches(msg, m.keyMap.Down) ||
			key.Matches(msg, m.keyMap.Up):
			m, cmd = m.inputFocusChange(msg)
			cmds = append(cmds, cmd)
		}

		switch msg.Type {
		case tea.KeyEsc:
			return m, message.TeaCmd(MsgClose{})
		default:
			// Handle all other key events
			m, cmd = m.focusedInputProcessKeyEvent(msg)
			cmds = append(cmds, cmd)
		}
	}

	m.viewport.SetContent(m.inputsView())

	return m, tea.Batch(cmds...)
}

func (m editModel) save(_ tea.Msg) (editModel, tea.Cmd) {
	for i := range m.inputs {
		switch i {
		case 0:
			m.host.Title = m.inputs[i].Value()
		case 1:
			m.host.Address = m.inputs[i].Value()
		case 2:
			m.host.Description = m.inputs[i].Value()
		case 3:
			m.host.LoginName = m.inputs[i].Value()
		case 4:
			m.host.RemotePort = m.inputs[i].Value()
		case 5:
			m.host.PrivateKeyPath = m.inputs[i].Value()
		}
	}

	_ = m.hostStorage.Save(m.host)
	// BUG: Когда мы отправляем hostlist.MsgRepoUpdated{}, компонент List неактивен и
	// следовательно не получает сообщения, смотри main
	return m, tea.Batch(TeaCmd(MsgClose{}), TeaCmd(hostlist.MsgRepoUpdated{}))
}

func (m editModel) focusedInputProcessKeyEvent(msg tea.Msg) (editModel, tea.Cmd) {
	var cmd tea.Cmd
	m.inputs[m.focusIndex], cmd = m.inputs[m.focusIndex].Update(msg)

	return m, cmd
}

func (m editModel) updateViewPort(msg tea.Msg) editModel {
	headerHeight := lipgloss.Height(m.headerView())
	helpMenuHeight := lipgloss.Height(m.helpView())

	if !m.ready {
		m.ready = true
		// m.viewport = viewport.New(m.size.Width, m.size.Height-headerHeight-helpMenuHeight)
		m.viewport = viewport.New(m.appState.Width, m.appState.Height-headerHeight-helpMenuHeight)
	} else if resizeMsg, ok := msg.(tea.WindowSizeMsg); ok {
		m.viewport.Width = resizeMsg.Width
		m.viewport.Height = resizeMsg.Height - headerHeight - helpMenuHeight
	}

	return m
}

func (m editModel) inputFocusChange(msg tea.Msg) (editModel, tea.Cmd) {
	var cmd tea.Cmd
	keyMsg := msg.(tea.KeyMsg)

	minFocusIndex := 0
	maxFocusIndex := len(m.inputs) - 1
	inputHeight := 0

	if len(m.inputs) > 0 {
		inputHeight = lipgloss.Height(m.inputsView()) / len(m.inputs)
	}

	// Update index of the focused element
	if key.Matches(keyMsg, m.keyMap.Up) && m.focusIndex > minFocusIndex { //nolint:gocritic // no need switch block here
		m.focusIndex--

		// Control viewport manually because height of input element is greater than one
		// therefore, we need to scroll several lines at once instead of just a single line.
		// Normally we don't need to handle scroll events, other than forward app messages to
		// the viewport: m.viewport, cmd = m.viewport.Update(msg)
		m.viewport.LineUp(inputHeight)
	} else if key.Matches(keyMsg, m.keyMap.Down) && m.focusIndex < maxFocusIndex {
		m.focusIndex++
		m.viewport.LineDown(inputHeight)
	} else {
		return m, nil
	}

	for i := 0; i <= len(m.inputs)-1; i++ {
		if i == m.focusIndex {
			// Set focused state
			cmd = m.inputs[i].Focus()
		} else {
			// Remove focused state
			m.inputs[i].Blur()
		}
	}

	return m, cmd
}

func (m editModel) inputsView() string {
	var b strings.Builder
	for i := range m.inputs {
		input := m.inputs[i]

		b.WriteString(input.View())
		if i < len(m.inputs) {
			b.WriteString("\n\n")
		}
	}

	return docStyle.Render(b.String())
}

func (m editModel) headerView() string {
	return titleStyle.Render("add a new host")
}

func (m editModel) helpView() string {
	return menuStyle.Render(m.help.View(m.keyMap))
}

func (m editModel) View() string {
	if !m.ready {
		// this should never happen, because Update method where we set "ready" to "true" triggers first
		return "Initializing..."
	}

	viewPortContent := m.viewport.View()
	return fmt.Sprintf("%s\n%s\n%s", m.headerView(), viewPortContent, m.helpView())
}
