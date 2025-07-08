// Package input implements generic UI input component.
package input

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/samber/lo"
)

// New - component which consists from input and label.
func New() *Input {
	inputModel := textinput.New()
	inputModel.Prompt = ""

	return &Input{
		Model:         inputModel,
		FocusedPrompt: "│ ",
		enabled:       true,
	}
}

// Input - input UI component.
type Input struct {
	textinput.Model

	label          string
	FocusedPrompt  string
	Tooltip        string
	Err            error
	enabled        bool
	displayTooltip bool
}

//nolint:revive // Init function is a part of tea component interface
func (l *Input) Init() tea.Cmd { return nil }

//nolint:revive // Update function is a part of tea component interface
func (l *Input) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	_, ok := msg.(tea.KeyMsg)
	if ok && !l.Enabled() {
		// If Input is disabled and it's a key message, then ignore it
		return l, nil
	}

	l.Model, cmd = l.Model.Update(msg)

	if l.Model.Validate != nil {
		l.Err = l.Model.Validate(l.Model.Value())
	}

	return l, cmd
}

//nolint:revive // View function is a part of tea component interface
func (l *Input) View() string {
	view := l.Model.View()

	if l.Focused() {
		view = focusedInputText.Render(view)
	} else if !l.Enabled() {
		view = greyedOutStyle.Render(view)
	}

	if l.displayTooltip && strings.TrimSpace(l.Tooltip) != "" {
		tooltip := lo.Ternary(l.Focused(), focusedStyle.Render(l.Tooltip), l.Tooltip)
		view = fmt.Sprintf("%s %s", tooltip, view)
	}

	return fmt.Sprintf("%s\n%s%s", l.labelView(), l.prompt(), view)
}

// Focus the Input if it's not disabled.
func (l *Input) Focus() tea.Cmd {
	if l.Enabled() {
		return l.Model.Focus()
	}

	return nil
}

func (l *Input) prompt() string {
	if l.Focused() {
		return focusedStyle.Render(l.FocusedPrompt)
	}

	return strings.Repeat(" ", utf8.RuneCountInString(l.FocusedPrompt))
}

func (l *Input) labelView() string {
	switch {
	case l.Err != nil:
		return l.prompt() + errorStyle.Render(l.Label())
	case l.Focused():
		return l.prompt() + focusedStyle.Render(l.Label())
	case !l.Enabled():
		return l.prompt() + greyedOutStyle.Render(l.Label())
	default:
		return l.prompt() + noStyle.Render(l.Label())
	}
}

// SetEnabled controls whether the component can be focused and changed.
func (l *Input) SetEnabled(isEnabled bool) {
	l.enabled = isEnabled
}

// Enabled returns component status - whether it can be changed or not.
func (l *Input) Enabled() bool {
	return l.enabled
}

// SetLabel sets the label of the Input.
func (l *Input) SetLabel(label string) {
	l.label = label
}

// Label returns the label value of the Input field.
func (l *Input) Label() string {
	return l.label
}

// SetDisplayTooltip manages tooltip text which is displayed in the beginning of the input field.
//
// Parameters:
// isDisplayed bool: a boolean value indicating whether the tooltip should be displayed.
func (l *Input) SetDisplayTooltip(isDisplayed bool) {
	l.displayTooltip = isDisplayed
}
