package theme

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/grafviktor/goto/internal/logger"
)

var currentTheme *Theme

// SetTheme sets the current application theme.
func SetTheme(theme *Theme) {
	currentTheme = theme
}

// GetTheme returns the current application theme.
func GetTheme() *Theme {
	if currentTheme == nil {
		currentTheme = DefaultTheme()
	}
	return currentTheme
}

// LoadTheme loads a theme from file or falls back to default.
func LoadTheme(configDir string) *Theme {
	themeFile := filepath.Join(configDir, "theme.json")

	theme, err := LoadThemeFromFile(themeFile)
	if err != nil {
		// Fall back to default theme
		theme = DefaultTheme()
		// Save the default theme to file for future use
		if saveErr := SaveThemeToFile(theme, themeFile); saveErr != nil {
			logger.Get().Error("Failed to save default theme: %v", saveErr)
		}
	}

	SetTheme(theme)
	return theme
}

// LoadThemeFromFile loads a theme from a JSON file.
func LoadThemeFromFile(filePath string) (*Theme, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read theme file: %w", err)
	}

	var theme Theme
	if err = json.Unmarshal(data, &theme); err != nil {
		return nil, fmt.Errorf("failed to parse theme file: %w", err)
	}

	theme.computeStyles()
	return &theme, nil
}

// SaveThemeToFile saves a theme to a JSON file.
func SaveThemeToFile(theme *Theme, filePath string) error {
	data, err := json.MarshalIndent(theme, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal theme: %w", err)
	}

	if err = os.WriteFile(filePath, data, 0o644); err != nil { //nolint:gosec // File permissions are sufficient
		return fmt.Errorf("failed to write theme file: %w", err)
	}

	return nil
}
