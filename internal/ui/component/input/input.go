// Package input implements generic UI input component.
package input

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"github.com/grafviktor/goto/internal/utils"
	"github.com/samber/lo"
)

// Input - input UI component.
type Input struct {
	textinput.Model

	label          string
	FocusedPrompt  string
	Tooltip        string
	Err            error
	enabled        bool
	displayTooltip bool
	styles         styles
}

// New - component which consists from input and label.
func New() *Input {
	inputModel := textinput.New()
	inputModel.Prompt = ""
	styles := defaultStyles()
	s := textinput.DefaultStyles(true)
	s.Focused.Placeholder = styles.textReadonly
	s.Focused.Text = styles.textFocused
	s.Cursor.Color = styles.cursor.GetForeground()
	inputModel.SetStyles(s)

	return &Input{
		Model:         inputModel,
		FocusedPrompt: "│ ",
		enabled:       true,
		styles:        styles,
	}
}

func (i *Input) Init() tea.Cmd { return nil }

func (i *Input) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	_, ok := msg.(tea.KeyPressMsg)
	if ok && !i.Enabled() {
		// If Input is disabled and it's a key message, then ignore it
		return i, nil
	}

	i.Model, cmd = i.Model.Update(msg)
	i.setPlaceholderWidth()

	if i.Model.Validate != nil {
		i.Err = i.Model.Validate(i.Model.Value())
	}

	return i, cmd
}

func (i *Input) SetValue(value string) {
	i.setPlaceholderWidth()
	i.Model.SetValue(value)
}

// TODO: Cover with tests
func (i *Input) setPlaceholderWidth() {
	// Bubbletea bug!
	// If not resize explicitly, there will be visible just a first letter
	// of placeholder text. See https://github.com/charmbracelet/bubbles/issues/779
	// See also: bubbles/v2@v2.0.0/textinput/textinput.go#748 (placeholderView function)
	// p := make([]rune, m.Width()+1)
	// copy(p, []rune(m.Placeholder))

	value := i.Model.Value()
	if utils.StringEmpty(&value) {
		i.SetWidth(len(i.Placeholder))
	} else {
		// If there is a value, then just reset width back to its initial state, which is 0, and let the component calculate it by itself.
		i.SetWidth(0)
	}
}

func (i *Input) View() tea.View {
	view := i.Model.View()

	switch {
	case i.Focused():
		view = i.styles.textFocused.Render(view)
	case !i.Enabled():
		view = i.styles.textReadonly.Render(view)
	default:
		view = i.styles.textNormal.Render(view)
	}

	if i.displayTooltip && strings.TrimSpace(i.Tooltip) != "" {
		tooltip := lo.Ternary(i.Focused(), i.styles.textFocused.Render(i.Tooltip), i.Tooltip)
		view = fmt.Sprintf("%s %s", tooltip, view)
	}

	viewContent := fmt.Sprintf("%s\n%s%s", i.labelView(), i.prompt(), view)
	return tea.NewView(viewContent)
}

// Focus the Input if it's not disabled.
func (i *Input) Focus() tea.Cmd {
	if i.Enabled() {
		return i.Model.Focus()
	}

	return nil
}

func (i *Input) prompt() string {
	if i.Focused() {
		return i.styles.textFocused.Render(i.FocusedPrompt)
	}

	return strings.Repeat(" ", utf8.RuneCountInString(i.FocusedPrompt))
}

func (i *Input) labelView() string {
	switch {
	case i.Err != nil:
		return i.prompt() + i.styles.inputError.Render(i.Label())
	case i.Focused():
		return i.prompt() + i.styles.inputFocused.Render(i.Label())
	case !i.Enabled():
		return i.prompt() + i.styles.textReadonly.Render(i.Label())
	default:
		return i.prompt() + i.styles.textNormal.Render(i.Label())
	}
}

// SetEnabled controls whether the component can be focused and changed.
func (i *Input) SetEnabled(isEnabled bool) {
	i.enabled = isEnabled
}

// Enabled returns component status - whether it can be changed or not.
func (i *Input) Enabled() bool {
	return i.enabled
}

// SetLabel sets the label of the Input.
func (i *Input) SetLabel(label string) {
	i.label = label
}

// Label returns the label value of the Input field.
func (i *Input) Label() string {
	return i.label
}

// SetDisplayTooltip manages tooltip text which is displayed in the beginning of the input field.
//
// Parameters:
// isDisplayed bool: a boolean value indicating whether the tooltip should be displayed.
func (i *Input) SetDisplayTooltip(isDisplayed bool) {
	i.displayTooltip = isDisplayed
}
