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
	theme, err := loadThemeFromFile(themeFile)
	if err != nil {
		theme = DefaultTheme()
		err = extractThemeFiles(filepath.Join(configDir, "themes"))
	}

	SetTheme(theme)
	return theme
}

// loadThemeFromFile loads a theme from a JSON file.
func loadThemeFromFile(filePath string) (*Theme, error) {
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

func extractThemeFiles(themesPath string) error {
	// 1. Write default theme to disk
	data, err := json.MarshalIndent(DefaultTheme(), "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal theme: %w", err)
	}
	saveThemeToFile(data, filepath.Join(themesPath, "default.json"))

	// 2. Extract embedded themes
	entries, err := resources.Themes.ReadDir("themes")
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		embeddedFSPath := filepath.Join("themes", entry.Name())
		data, err := resources.Themes.ReadFile(embeddedFSPath)
		if err != nil {
			continue
		}

		saveThemeToFile(data, filepath.Join(themesPath, entry.Name()))
	}

	return nil
}

func saveThemeToFile(theme []byte, filePath string) error {
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0o755); err != nil { //nolint:gosec // Folder permissions are sufficient
		return fmt.Errorf("failed to create directory: %w", err)
	}

	if err := os.WriteFile(filePath, theme, 0o644); err != nil { //nolint:gosec // File permissions are sufficient
		return fmt.Errorf("failed to write theme file: %w", err)
	}

	return nil
}
