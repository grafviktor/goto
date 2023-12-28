// Package edithost contains UI components for editing host model attributes.
package edithost

import (
	"context"
	"fmt"
	"strconv"
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
	"github.com/grafviktor/goto/internal/utils"
)

// Size struct is used by terminal resize event.
type Size struct {
	Width  int
	Height int
}

type (
	// MsgClose triggers when users exits from edit form without saving results.
	MsgClose struct{}
	// MsgSave triggers when users saves results.
	MsgSave struct{}
)

const (
	inputTitle int = iota
	inputAddress
	inputDescription
	inputLogin
	inputNetworkPort
	inputIdentityFile
)

// ItemID is a key to extract item id from application context.
var ItemID = struct{}{}

type logger interface {
	Debug(format string, args ...any)
}

func notEmptyValidator(s string) error {
	if utils.StringEmpty(s) {
		return fmt.Errorf("value is required")
	}

	return nil
}

func networkPortValidator(s string) error {
	if utils.StringEmpty(s) {
		return nil
	}

	auto := 0 // 0 is used to autodetect base, see strconv.ParseUint
	maxLengthBit := 16
	if num, err := strconv.ParseUint(s, auto, maxLengthBit); err != nil || num < 1 {
		return fmt.Errorf("network port must be a number which is less than 65 535")
	}

	return nil
}

// New - returns new edit host form.
func New(ctx context.Context, storage storage.HostStorage, state *state.ApplicationState, log logger) editModel {
	initialFocusedInput := inputTitle
	isNewHost := false

	// if we can't cast host id to int, that means we're adding a new host. Ignoring the error
	hostID, _ := ctx.Value(ItemID).(int)
	host, err := storage.Get(hostID)
	if err != nil {
		host = model.Host{}
		initialFocusedInput = inputAddress
		isNewHost = true
	}

	m := editModel{
		inputs:       make([]labeledInput, 6),
		hostStorage:  storage,
		host:         host,
		help:         help.New(),
		keyMap:       keys,
		appState:     state,
		logger:       log,
		focusedInput: initialFocusedInput,
		isNewHost:    isNewHost,
	}

	var t labeledInput
	for i := range m.inputs {
		t = NewLabelInput()
		t.Cursor.Style = cursorStyle

		switch i {
		case inputTitle:
			t.Label = "Title"
			t.SetValue(host.Title)
			t.Placeholder = "title"
			t.Validate = notEmptyValidator
		case inputAddress:
			t.Label = "IP Address or Hostname"
			t.CharLimit = 128
			t.SetValue(host.Address)
			t.Placeholder = "address"
			t.Validate = notEmptyValidator
		case inputDescription:
			t.Label = "Description"
			t.CharLimit = 512
			t.Placeholder = "n/a"
			t.SetValue(host.Description)
		case inputLogin:
			t.Label = "Login"
			t.CharLimit = 128
			t.Placeholder = fmt.Sprintf("default: %s", utils.CurrentUsername())
			t.SetValue(host.LoginName)
		case inputNetworkPort:
			t.Label = "Network port"
			t.CharLimit = 5
			t.Placeholder = "default: 22"
			t.SetValue(host.RemotePort)
			t.Validate = networkPortValidator
		case inputIdentityFile:
			t.Label = "Identity file path"
			t.CharLimit = 512
			t.Placeholder = "default: $HOME/.ssh/id_rsa"
			t.SetValue(host.PrivateKeyPath)
		}

		m.inputs[i] = t
	}

	m.inputs[m.focusedInput].Focus()

	return m
}

type editModel struct {
	keyMap       keyMap
	hostStorage  storage.HostStorage
	focusedInput int
	inputs       []labeledInput
	host         model.Host
	isNewHost    bool
	viewport     viewport.Model
	help         help.Model
	ready        bool
	appState     *state.ApplicationState
	logger       logger
}

func (m editModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m editModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// this message never comes through on windows. Sending it from init_win.go
	if msg, ok := msg.(tea.WindowSizeMsg); ok {
		m.logger.Debug("Resizing edit host viewport: %d %d", msg.Width, msg.Height)
	}

	var cmd tea.Cmd
	var cmds []tea.Cmd

	// forward keyboard events to help menu component
	m.help, cmd = m.help.Update(msg)
	cmds = append(cmds, cmd)

	// create or Update viewport
	m = m.updateViewPort(msg)

	if msg, ok := msg.(tea.KeyMsg); ok {
		switch {
		case key.Matches(msg, m.keyMap.Save):
			m, cmd = m.save(msg)
			cmds = append(cmds, cmd)
		case key.Matches(msg, m.keyMap.Down) || key.Matches(msg, m.keyMap.Up):
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
		case inputTitle:
			m.host.Title = m.inputs[i].Value()
		case inputAddress:
			m.host.Address = m.inputs[i].Value()
		case inputDescription:
			m.host.Description = m.inputs[i].Value()
		case inputLogin:
			m.host.LoginName = m.inputs[i].Value()
		case inputNetworkPort:
			m.host.RemotePort = m.inputs[i].Value()
		case inputIdentityFile:
			m.host.PrivateKeyPath = m.inputs[i].Value()
		}
	}

	_ = m.hostStorage.Save(m.host)
	return m, tea.Batch(
		message.TeaCmd(MsgClose{}),
		message.TeaCmd(hostlist.MsgRepoUpdated{}),
	)
}

func (m editModel) copyAddressToTitle() {
	newValue := m.inputs[inputAddress].Value()

	// Temprorary remove input validator.
	// It's necessary, because input.SetValue(...) invokes Validate function,
	// if the input contains invalid value, Validate function returns error and
	// rejects new value. That leads to a problem - when user removes all symbols
	// from address input, title input still preserves the very last letter.
	// A better way would be to use own validation logic instead of relying
	// on input.Validate.
	validator := m.inputs[inputTitle].Validate
	m.inputs[inputTitle].Validate = nil
	m.inputs[inputTitle].SetValue(newValue)
	m.inputs[inputTitle].SetCursor(len(newValue))
	m.inputs[inputTitle].Validate = validator
	m.inputs[inputTitle].Err = m.inputs[inputTitle].Validate(newValue)
}

func (m editModel) focusedInputProcessKeyEvent(msg tea.Msg) (editModel, tea.Cmd) {
	var cmd tea.Cmd
	var shouldUpdateTitle bool

	// Decide if we need to propagate hostname to title
	if m.focusedInput == inputAddress {
		addressEqualsTitle := m.inputs[inputAddress].Value() == m.inputs[inputTitle].Value()
		shouldUpdateTitle = m.isNewHost && addressEqualsTitle
	}

	// Update focused input
	m.inputs[m.focusedInput], cmd = m.inputs[m.focusedInput].Update(msg)

	// Then, update title if we should
	if shouldUpdateTitle {
		m.copyAddressToTitle()
	}

	return m, cmd
}

func (m editModel) updateViewPort(msg tea.Msg) editModel {
	headerHeight := lipgloss.Height(m.headerView())
	helpMenuHeight := lipgloss.Height(m.helpView())

	if !m.ready {
		m.ready = true
		m.viewport = viewport.New(m.appState.Width, m.appState.Height-headerHeight-helpMenuHeight)
	} else if resizeMsg, ok := msg.(tea.WindowSizeMsg); ok {
		m.viewport.Width = resizeMsg.Width
		m.viewport.Height = resizeMsg.Height - headerHeight - helpMenuHeight
		m.logger.Debug("Resizing edit host viewport: %d %d", m.viewport.Width, m.viewport.Height)
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
		// Control viewport manually because height of input element is greater than one
		// therefore, we need to scroll several lines at once instead of just a single line.
		// Normally we don't need to handle scroll events, other than forward app messages to
		// the viewport: m.viewport, cmd = m.viewport.Update(msg)
		inputHeight = lipgloss.Height(m.inputsView()) / len(m.inputs)
	}

	// Update index of the focused element
	if key.Matches(keyMsg, m.keyMap.Up) && m.focusedInput > minFocusIndex { //nolint:gocritic // no need switch block here
		m.focusedInput--
		m.viewport.LineUp(inputHeight)
	} else if key.Matches(keyMsg, m.keyMap.Down) && m.focusedInput < maxFocusIndex {
		m.focusedInput++
		m.viewport.LineDown(inputHeight)
	} else {
		return m, nil
	}

	for i := 0; i <= len(m.inputs)-1; i++ {
		if m.inputs[i].Validate != nil {
			m.inputs[i].Err = m.inputs[i].Validate(m.inputs[i].Value())
		}

		if i == m.focusedInput {
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
	return titleStyle.Render("edit host")
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
