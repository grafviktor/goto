// Package hostedit contains UI components for editing host model attributes.
package hostedit

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/samber/lo"

	hostModel "github.com/grafviktor/goto/internal/model/host"
	"github.com/grafviktor/goto/internal/model/sshconfig"
	"github.com/grafviktor/goto/internal/state"
	"github.com/grafviktor/goto/internal/storage"
	"github.com/grafviktor/goto/internal/ui/component/input"
	"github.com/grafviktor/goto/internal/ui/message"
	"github.com/grafviktor/goto/internal/utils"
)

// Size struct is used by terminal resize event.
type Size struct {
	Width  int
	Height int
}

type (
	// MsgSave triggers when users saves results.
	MsgSave struct{}
	// debouncedMessage is used to trigger side effects. For instance dispatch RunProcessSSHLoadConfig
	// which reads host config from ~/.ssh/config file.
	debouncedMessage struct {
		wrappedMsg  tea.Msg
		debounceTag int
	}
)

const (
	inputTitle int = iota
	inputAddress
	inputDescription
	inputGroup
	inputLogin
	inputNetworkPort
	inputIdentityFile
)

type itemID struct{}

var (
	// ItemID is a key to extract item id from application context.
	ItemID       = itemID{}
	defaultTitle = "host details"
	debounceTime = time.Millisecond * 300
)

type iLogger interface {
	Error(format string, args ...any)
	Debug(format string, args ...any)
	Info(format string, args ...any)
}

func notEmptyValidator(s string) error {
	if utils.StringEmpty(&s) {
		return errors.New("value is required")
	}

	return nil
}

func networkPortValidator(s string) error {
	if utils.StringEmpty(&s) {
		return nil
	}

	auto := 0 // 0 is used to autodetect base, see strconv.ParseUint
	maxLengthBit := 16
	if num, err := strconv.ParseUint(s, auto, maxLengthBit); err != nil || num < 1 {
		return errors.New("network port must be a number which is less than 65,535")
	}

	return nil
}

func getKeyMap(host hostModel.Host, focusedInput int) keyMap {
	switch {
	case host.IsReadOnly():
		keys.CopyInputValue.SetEnabled(false)
		keys.Save.SetEnabled(false)
		keys.Up.SetEnabled(false)
		keys.Down.SetEnabled(false)
		keys.Discard.SetHelp("esc", "close")
	case focusedInput == inputTitle || focusedInput == inputAddress:
		keys.CopyInputValue.SetEnabled(true)
		keys.Save.SetEnabled(true)
		keys.Up.SetEnabled(true)
		keys.Down.SetEnabled(true)
		keys.Discard.SetHelp("esc", "discard")
	default:
		keys.CopyInputValue.SetEnabled(false)
		keys.Save.SetEnabled(true)
		keys.Up.SetEnabled(true)
		keys.Down.SetEnabled(true)
		keys.Discard.SetHelp("esc", "discard")
	}

	return keys
}

type EditModel struct {
	appState     *state.State
	focusedInput int
	help         help.Model
	host         hostModelWrapper
	hostStorage  storage.HostStorage
	inputs       []input.Input
	isNewHost    bool
	keyMap       keyMap
	logger       iLogger
	ready        bool
	title        string
	viewport     viewport.Model
	debounceTag  int
	styles       styles
}

// New - returns new edit host form.
func New(ctx context.Context, storage storage.HostStorage, state *state.State, log iLogger) *EditModel {
	initialFocusedInput := inputTitle

	// If we can't cast host id to int, that means we're adding a new host. Ignore the error
	hostID, _ := ctx.Value(ItemID).(int)
	host, hostNotFoundErr := storage.Get(hostID)
	if hostNotFoundErr != nil {
		// Logger should notify that this is a new host
		host = hostModel.Host{Group: state.Group}
	}
	host.SSHHostConfig = sshconfig.StubConfig()

	m := EditModel{
		inputs:       make([]input.Input, 7), //nolint:mnd // Quantity of input components is 7
		hostStorage:  storage,
		host:         wrap(&host),
		help:         help.New(),
		keyMap:       getKeyMap(host, initialFocusedInput),
		appState:     state,
		logger:       log,
		focusedInput: initialFocusedInput,
		title:        defaultTitle,
		// This variable is for optimization. By introducing it, we can avoid unnecessary database reads
		// every time we change values which depend on each other, for instance: "Title" and "Address".
		// Use text search and see where 'isNewHost' is used.
		isNewHost: hostNotFoundErr != nil,
		styles:    defaultStyles(),
	}

	m.help.Styles = m.styles.help

	var t input.Input
	for i := range m.inputs {
		t = *input.New()
		t.Cursor.Style = m.styles.cursor

		switch i {
		case inputTitle:
			t.SetLabel("Title")
			t.SetValue(host.Title)
			t.Validate = notEmptyValidator
		case inputAddress:
			t.SetLabel("Host")
			t.CharLimit = 128
			t.SetValue(host.Address)
			t.Validate = notEmptyValidator
			t.Tooltip = "ssh"
		case inputDescription:
			t.SetLabel("Description")
			t.CharLimit = 512
			t.SetValue(host.Description)
		case inputGroup:
			t.SetLabel("Group")
			t.CharLimit = 512
			t.SetValue(host.Group)
		case inputLogin:
			t.SetLabel("Login")
			t.CharLimit = 128
			t.SetValue(host.LoginName)
		case inputNetworkPort:
			t.SetLabel("Network Port")
			t.CharLimit = 5
			t.SetValue(host.RemotePort)
			t.Validate = networkPortValidator
		case inputIdentityFile:
			t.SetLabel("Identity File")
			t.CharLimit = 512
			t.SetValue(host.IdentityFilePath)
		}

		m.inputs[i] = t
	}

	// Though updateInputFields will automatically be called once ssh config is loaded,
	// that will not happen when we create a new host. Thus calling it manually.
	m.updateInputFields()
	m.inputs[m.focusedInput].Focus()

	return &m
}

func (m *EditModel) Init() tea.Cmd { return nil }

func (m *EditModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// This message never comes through automatically on Windows OS, we send it from init_win.go.
		m.updateViewPort(msg)
	case tea.KeyMsg:
		cmd = m.handleKeyboardEvent(msg)
		m.viewport.SetContent(m.inputsView())
	case debouncedMessage:
		cmd = m.handleDebouncedMessage(msg)
	case message.HostSSHConfigLoaded:
		m.host.SSHHostConfig = &msg.Config
		m.updateInputFields()
		m.viewport.SetContent(m.inputsView())
	case message.HideUINotification:
		if msg.ComponentName == "hostedit" {
			m.logger.Debug("[UI] Hide notification message")
			m.title = defaultTitle
		}

		return m, nil
	}

	return m, cmd
}

func (m *EditModel) View() string {
	if !m.ready {
		// Create viewport, ideally this call should be located in init function,
		// but this function does not trigger for child components
		m.updateViewPort(nil)
	}

	viewPortContent := m.viewport.View()
	return fmt.Sprintf("%s\n%s\n%s", m.headerView(), viewPortContent, m.helpView())
}

func (m *EditModel) handleKeyboardEvent(msg tea.KeyMsg) tea.Cmd {
	// If title displays an error, due to an incorrect title for instance
	// once user presses any button, we should reset it to default value
	m.title = defaultTitle

	switch {
	case key.Matches(msg, m.keyMap.Discard):
		m.logger.Info("[UI] Discard changes for host id: %v", m.host.ID)
		return message.TeaCmd(message.CloseViewHostEdit{})
	case m.host.IsReadOnly():
		m.logger.Debug("[UI] Received a key event. Cannot modify a readonly host.")
		return m.displayNotificationMsg("host loaded from SSH config is readonly")
	case key.Matches(msg, m.keyMap.Save):
		m.logger.Info("[UI] Save changes for host id: %v", m.host.ID)
		return m.save(msg)
	case key.Matches(msg, m.keyMap.CopyInputValue):
		m.handleCopyInputValueShortcut()
		return nil
	case key.Matches(msg, m.keyMap.Down) || key.Matches(msg, m.keyMap.Up):
		return m.inputFocusChange(msg)
	default:
		// Handle all other key events
		cmd := m.focusedInputProcessKeyEvent(msg)
		if m.focusedInput == inputAddress || m.focusedInput == inputTitle {
			// This statement is required as user may want to copy title to address,
			// if Host field contains a custom command, ssh options inputs
			// should be disabled.
			m.updateInputFields()
		}

		return cmd
	}
}

func (m *EditModel) handleDebouncedMessage(msg debouncedMessage) tea.Cmd {
	// This function debounces a tea.Message. In order to find the last message from a list of duplicate messages
	// debounceTag is used. Every time a tea.Tick message is dispatched, debounceTag is incremented. Then, when
	// tea.Tick message triggers by timer (by debounceTime) it compares its own debounceTag with the model's
	// debounceTag and only triggers when they're equal. That guarantees that only last message will be handled.
	m.debounceTag++

	return tea.Tick(debounceTime, func(_ time.Time) tea.Msg {
		// Need to decrement the model's debounce tag before comparing. This simply relates to order of operations.
		if msg.debounceTag == m.debounceTag-1 {
			// Only the last message from messages dispatched within a certain interval will be handled.
			return msg.wrappedMsg
		}

		return nil
	})
}

func (m *EditModel) save(_ tea.Msg) tea.Cmd {
	for i := range m.inputs {
		if m.inputs[i].Validate != nil {
			if err := m.inputs[i].Validate(m.inputs[i].Value()); err != nil {
				m.logger.Info(
					"[UI] Cannot save host with id %v. Reason: '%s' is not valid, %s",
					m.host.ID,
					m.inputs[i].Label(),
					err.Error(),
				)
				m.inputs[i].Err = err
				m.title = fmt.Sprintf("%s is not valid", m.inputs[i].Label())

				return nil
			}
		}

		if i == inputGroup {
			// When create a new host within specific group, the app pre-populates group value
			// however, there is a chance that the group field will never be focused by user
			// and the value will never be copied to the model. This is a workaround for this issue.
			// Explicitly set group value to the model.
			groupName := m.inputs[i].Value()
			m.host.setHostAttributeByIndex(i, strings.TrimSpace(groupName))
		}
	}

	var cmd tea.Cmd
	host, err := m.hostStorage.Save(m.host.unwrap())
	if err != nil {
		m.logger.Error("[UI] Cannot save host with id %v. Reason: %s", m.host.ID, err.Error())
		cmd = message.TeaCmd(message.ErrorOccurred{Err: err})
	} else {
		cmd = lo.Ternary(m.isNewHost,
			message.TeaCmd(message.HostCreated{Host: host}),
			message.TeaCmd(message.HostUpdated{Host: host}))
	}

	return tea.Sequence(
		message.TeaCmd(message.CloseViewHostEdit{}),
		// Order matters here! That's why we use tea.Sequence instead of tea.Batch.
		message.TeaCmd(message.HostSelected{HostID: host.ID}),
		cmd,
	)
}

func (m *EditModel) copyInputValueFromTo(sourceInput, destinationInput int) {
	newValue := m.inputs[sourceInput].Value()

	// Temporary remove input validator.
	// It's necessary, because input.SetValue(...) invokes Validate function,
	// if the input contains invalid value, Validate function returns error and
	// rejects new value. That leads to a problem - when user removes all symbols
	// from address input, title input still preserves the very last letter.
	// A better way would be to use own validation logic instead of relying
	// on input.Validate.
	validator := m.inputs[destinationInput].Validate
	m.inputs[destinationInput].Validate = nil
	m.inputs[destinationInput].SetValue(newValue)
	m.inputs[destinationInput].SetCursor(len(newValue))
	m.inputs[destinationInput].Validate = validator
	m.inputs[destinationInput].Err = m.inputs[destinationInput].Validate(newValue)
	m.logger.Debug(
		"[UI] Copy '%s' value to '%s', new value = %s",
		m.inputs[sourceInput].Label(),
		m.inputs[destinationInput].Label(),
		newValue,
	)

	// Update the model as well
	m.host.setHostAttributeByIndex(destinationInput, newValue)
}

func (m *EditModel) focusedInputProcessKeyEvent(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	var shouldUpdateTitle bool
	previousValue := m.inputs[m.focusedInput].Value()

	// Decide if we need to propagate hostname to title.
	// Note, that we should make this decision BEFORE updating focused input
	if m.focusedInput == inputTitle {
		addressEqualsTitle := m.inputs[inputTitle].Value() == m.inputs[inputAddress].Value()

		// If there wouldn't be 'm.isNewHost' variable we would have to query database for every key event
		if m.isNewHost && addressEqualsTitle {
			// If host doesn't exist in the repo and title equals address
			// we should copy text from address to title.
			shouldUpdateTitle = true
		}
	}

	// Update focused input
	m.inputs[m.focusedInput].Update(msg)

	// Then, update title if we should
	if shouldUpdateTitle {
		m.copyInputValueFromTo(inputTitle, inputAddress)
	}

	// When change UI field, update the model as well. We need to update the model on this stage
	// because we need to know its details(such as username, port number, ssh key) when request SSH config.
	// We assign value to the model here, then we read it again in "updateInputFields" function. That
	// should be simplified, we only need to assign value to the model when we save it.
	m.host.setHostAttributeByIndex(m.focusedInput, m.inputs[m.focusedInput].Value())

	// If update address or title field
	if lo.Contains([]int{inputTitle, inputAddress}, m.focusedInput) {
		currentValue := m.inputs[inputAddress].Value()

		// And value changed
		if previousValue != currentValue {
			// Load SSH config for the specified hostname
			cmd = message.TeaCmd(debouncedMessage{
				wrappedMsg:  message.RunProcessSSHLoadConfig{Host: *m.host.Host},
				debounceTag: m.debounceTag, // See the comments in debouncedMessage definition.
			})
		}
	}

	return cmd
}

func (m *EditModel) updateViewPort(msg tea.Msg) {
	headerHeight := lipgloss.Height(m.headerView())
	helpMenuHeight := lipgloss.Height(m.helpView())

	if !m.ready {
		m.ready = true
		m.viewport = viewport.New(m.appState.Width, m.appState.Height-headerHeight-helpMenuHeight)
		m.viewport.SetContent(m.inputsView())
	} else if resizeMsg, ok := msg.(tea.WindowSizeMsg); ok {
		m.viewport.Width = resizeMsg.Width
		m.viewport.Height = resizeMsg.Height - headerHeight - helpMenuHeight
		m.logger.Debug("[UI] Set edit host viewport size: %d %d", m.viewport.Width, m.viewport.Height)
	}
}

func (m *EditModel) inputFocusChange(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd
	keyMsg, _ := msg.(tea.KeyMsg)

	enabledInputs := lo.Filter(m.inputs, func(i input.Input, _ int) bool {
		return i.Enabled()
	})

	minFocusIndex := 0
	// maxFocusIndex is equal to number of inputs minus the number
	// of disabled inputs. This works based on an assumption that
	// all disabled inputs will be in the bottom of the hostlist.
	maxFocusIndex := len(enabledInputs) - 1
	inputHeight := 0

	if len(m.inputs) > 0 {
		// Control viewport manually because height of input element is greater than one
		// therefore, we need to scroll several lines at once instead of just a single line.
		// Normally we don't need to handle scroll events, other than forward app messages to
		// the viewport: m.viewport, cmd = m.viewport.Update(msg)
		inputHeight = lipgloss.Height(m.inputsView()) / len(m.inputs)
	}

	// Update index of the focused element
	if key.Matches(keyMsg, m.keyMap.Up) && m.focusedInput > minFocusIndex { //nolint:gocritic // readable without switch
		m.focusedInput--
		m.viewport.LineUp(inputHeight)
	} else if key.Matches(keyMsg, m.keyMap.Down) && m.focusedInput < maxFocusIndex {
		m.focusedInput++
		m.viewport.LineDown(inputHeight)
	} else {
		m.logger.Debug("[UI] Reached first or last selectable input field: %d", m.focusedInput)
		return nil
	}

	// Should be extracted to "Validate" function
	for i := range m.inputs {
		if m.inputs[i].Validate != nil {
			m.inputs[i].Err = m.inputs[i].Validate(m.inputs[i].Value())
			m.logger.Debug("[UI] Input '%v' is valid: %v", m.inputs[i].Label(), m.inputs[i].Err == nil)
		}

		if i == m.focusedInput {
			// KeyMap depends on focused input - when address is focused, we allow
			// a user to copy address value to title.
			m.keyMap = getKeyMap(m.host.unwrap(), i)
			m.logger.Debug("[UI] Focus input: '%s'", m.inputs[i].Label())

			// Set focused state
			cmds = append(cmds, m.inputs[i].Focus())
		} else {
			// Remove focused state
			m.inputs[i].Blur()
		}
	}

	return tea.Batch(cmds...)
}

func (m *EditModel) handleCopyInputValueShortcut() {
	// Allow a user to copy values between address and title,
	// because the chances are that these two inputs will have
	// the same values.
	switch m.focusedInput {
	case inputTitle:
		m.copyInputValueFromTo(m.focusedInput, inputAddress)
	case inputAddress:
		m.copyInputValueFromTo(m.focusedInput, inputTitle)
	}
}

func (m *EditModel) updateInputFields() {
	if m.host.IsReadOnly() {
		m.handleReadonlyHost()
	} else {
		m.handleEditableHost()
	}
}

func (m *EditModel) handleReadonlyHost() {
	m.logger.Debug("[UI] Update input components. All parameters are disabled.")
	lo.ForEach(m.inputs, func(_ input.Input, n int) {
		m.inputs[n].PlaceholderStyle = m.styles.textReadonly
		m.inputs[n].Placeholder = fmt.Sprintf("%s: %s", "readonly", m.host.getHostAttributeValueByIndex(n))
		m.inputs[n].SetValue("")
		m.inputs[n].SetEnabled(false)
	})
}

func (m *EditModel) handleEditableHost() {
	customConnectString := m.host.IsUserDefinedSSHCommand()
	m.logger.Debug("[UI] Update input components. Additional SSH parameters disabled: %v", customConnectString)

	prefix := lo.Ternary(customConnectString, "readonly", "default")
	m.inputs[inputTitle].Placeholder = "*required*"
	m.inputs[inputAddress].Placeholder = "*required*"
	m.inputs[inputGroup].Placeholder = "n/a"
	m.inputs[inputDescription].Placeholder = "n/a"
	m.inputs[inputLogin].Placeholder = fmt.Sprintf("%s: %s", prefix, m.host.SSHHostConfig.User)
	m.inputs[inputNetworkPort].Placeholder = fmt.Sprintf("%s: %s", prefix, m.host.SSHHostConfig.Port)
	m.inputs[inputIdentityFile].Placeholder = fmt.Sprintf("%s: %s", prefix, m.host.SSHHostConfig.IdentityFile)

	hostInputLabel := lo.Ternary(customConnectString, "Command", "Host")
	m.inputs[inputAddress].SetLabel(hostInputLabel)
	m.inputs[inputAddress].SetDisplayTooltip(customConnectString)

	// Enable or disable inputs based on the custom connect string //

	// Get input fields by pointer to update their state
	sshParamsInputFields := []*input.Input{
		&m.inputs[inputLogin],
		&m.inputs[inputNetworkPort],
		&m.inputs[inputIdentityFile],
	}

	lo.ForEach(sshParamsInputFields, func(i *input.Input, _ int) {
		i.SetEnabled(!customConnectString)
	})

	lo.ForEach(m.inputs, func(_ input.Input, n int) {
		if m.inputs[n].Enabled() {
			m.inputs[n].SetValue(m.host.getHostAttributeValueByIndex(n))
		} else {
			m.inputs[n].SetValue("")
		}
	})
}

func (m *EditModel) inputsView() string {
	var b strings.Builder
	for i := range m.inputs {
		b.WriteString(m.inputs[i].View())
		if i < len(m.inputs) {
			b.WriteString("\n\n")
		}
	}

	return m.styles.componentMargins.Render(b.String())
}

func (m *EditModel) headerView() string {
	return m.styles.title.Render(m.title)
}

func (m *EditModel) helpView() string {
	return m.styles.textReadonly.Render(m.help.View(m.keyMap))
}

func (m *EditModel) SetTitle(title string) {
	m.title = title
}

func (m *EditModel) displayNotificationMsg(msg string) tea.Cmd {
	if utils.StringEmpty(&msg) {
		return nil
	}

	m.logger.Debug("[UI] Notification message: %s", msg)
	return message.DisplayNotification("hostedit", msg, m)
}
