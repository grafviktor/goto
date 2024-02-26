package hostedit

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// NewLabeledInput - component which consists from input and label.
func NewLabeledInput() labeledInput {
	inputModel := textinput.New()
	inputModel.Prompt = ""

	return labeledInput{
		Model:         inputModel,
		FocusedPrompt: "│ ",
	}
}

type labeledInput struct {
	textinput.Model
	Label         string
	FocusedPrompt string
	Err           error
}

func (l labeledInput) Update(msg tea.Msg) (labeledInput, tea.Cmd) {
	var cmd tea.Cmd

	l.Model, cmd = l.Model.Update(msg)

	if l.Model.Validate != nil {
		l.Err = l.Model.Validate(l.Model.Value())
	}

	return l, cmd
}

func (l labeledInput) prompt() string {
	if l.Focused() {
		return focusedStyle.Render(l.FocusedPrompt)
	}

	return strings.Repeat(" ", utf8.RuneCountInString(l.FocusedPrompt))
}

func (l labeledInput) labelView() string {
	if l.Err != nil {
		return l.prompt() + errorStyle.Render(l.Label)
	} else if l.Focused() {
		return l.prompt() + focusedStyle.Render(l.Label)
	}

	return l.prompt() + noStyle.Render(l.Label)
}

func (l labeledInput) View() string {
	var view string
	if l.Focused() {
		view = focusedInputText.Render(l.Model.View())
	} else {
		view = l.Model.View()
	}

	return fmt.Sprintf("%s\n%s%s", l.labelView(), l.prompt(), view)
}
