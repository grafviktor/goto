package input

import (
	"github.com/grafviktor/goto/internal/ui/theme"
)

var (
	themeSettings = theme.GetTheme().Styles.Input

	styleInputError   = themeSettings.InputError
	styleInputFocused = themeSettings.InputFocused
	styleText         = themeSettings.TextNormal
	styleTextFocused  = themeSettings.TextFocused
	styleTextReadonly = themeSettings.TextReadonly
)
