// Package input implements generic UI input component.
package input

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// New - component which consists from input and label.
func New() *Input {
	inputModel := textinput.New()
	inputModel.Prompt = ""

	return &Input{
		Model:         inputModel,
		FocusedPrompt: "│ ",
	}
}

// Input - input UI component.
type Input struct {
	textinput.Model
	Label         string
	FocusedPrompt string
	Err           error
}

//nolint:revive // Init function is a part of tea component interface
func (l *Input) Init() tea.Cmd { return nil }

//nolint:revive // Update function is a part of tea component interface
func (l *Input) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	l.Model, cmd = l.Model.Update(msg)

	if l.Model.Validate != nil {
		l.Err = l.Model.Validate(l.Model.Value())
	}

	return l, cmd
}

//nolint:revive // View function is a part of tea component interface
func (l *Input) View() string {
	var view string
	if l.Focused() {
		view = focusedInputText.Render(l.Model.View())
	} else {
		view = l.Model.View()
	}

	return fmt.Sprintf("%s\n%s%s", l.labelView(), l.prompt(), view)
}

func (l *Input) prompt() string {
	if l.Focused() {
		return focusedStyle.Render(l.FocusedPrompt)
	}

	return strings.Repeat(" ", utf8.RuneCountInString(l.FocusedPrompt))
}

func (l *Input) labelView() string {
	if l.Err != nil {
		return l.prompt() + errorStyle.Render(l.Label)
	} else if l.Focused() {
		return l.prompt() + focusedStyle.Render(l.Label)
	}

	return l.prompt() + noStyle.Render(l.Label)
}
