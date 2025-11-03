package theme

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/grafviktor/goto/internal/resources"
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
func LoadTheme(configDir, themeName string) *Theme {
	themeFile := filepath.Join(configDir, "themes", themeName+".json")

	theme, err := LoadThemeFromFile(themeFile)
	if err != nil {
		// themeFiles, _ := resources.Themes.ReadDir("themes")
		// for _, themeFile := range themeFiles {
		// 	logger.Get().Debug("Available embedded theme: %s", themeFile.Name())
		// }
		// Fall back to default theme
		theme = DefaultTheme()
		// Save the default theme to file for future use
		extractThemeFiles(filepath.Join(configDir, "themes"))
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

func extractThemeFiles(filePath string) error {
	entries, err := resources.Themes.ReadDir("themes")
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		if !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		data, err := resources.Themes.ReadFile(entry.Name())
		if err != nil {
			return err
		}

		// Write to disk
		// outPath := filepath.Join(targetDir, entry.Name())
		// if err := os.WriteFile(outPath, data, 0o644); err != nil {
		// 	return err
		// }
	}

	return nil
}

// func SaveThemeToFile(theme *Theme, filePath string) error {
// 	dir := filepath.Dir(filePath)
// 	if err := os.MkdirAll(dir, 0o755); err != nil { //nolint:gosec // Folder permissions are sufficient
// 		return fmt.Errorf("failed to create directory: %w", err)
// 	}

// 	data, err := json.MarshalIndent(theme, "", "  ")
// 	if err != nil {
// 		return fmt.Errorf("failed to marshal theme: %w", err)
// 	}

// 	if err = os.WriteFile(filePath, data, 0o644); err != nil { //nolint:gosec // File permissions are sufficient
// 		return fmt.Errorf("failed to write theme file: %w", err)
// 	}

// 	return nil
// }

// // SaveThemeToFile saves a theme to a JSON file.
// func SaveThemeToFile(theme *Theme, filePath string) error {
// 	dir := filepath.Dir(filePath)
// 	if err := os.MkdirAll(dir, 0o755); err != nil { //nolint:gosec // Folder permissions are sufficient
// 		return fmt.Errorf("failed to create directory: %w", err)
// 	}

// 	data, err := json.MarshalIndent(theme, "", "  ")
// 	if err != nil {
// 		return fmt.Errorf("failed to marshal theme: %w", err)
// 	}

// 	if err = os.WriteFile(filePath, data, 0o644); err != nil { //nolint:gosec // File permissions are sufficient
// 		return fmt.Errorf("failed to write theme file: %w", err)
// 	}

// 	return nil
// }
