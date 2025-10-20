package input

import (
	"github.com/grafviktor/goto/internal/ui/theme"
)

func GetStyles() *theme.InputStyles {
	t := theme.GetTheme()
	return &t.Styles.Input
}

var (
	styleInputFocused = GetStyles().InputFocused
	styleInputError   = GetStyles().InputError
	styleTextFocused  = GetStyles().TextFocused
	styleTextReadonly = GetStyles().TextReadonly
	styleText         = GetStyles().TextNormal
)
