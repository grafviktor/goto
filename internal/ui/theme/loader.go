package theme

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/grafviktor/goto/internal/resources"
	"github.com/grafviktor/goto/internal/utils"
)

type loggerInterface interface {
	Info(format string, args ...any)
	Error(format string, args ...any)
	Debug(format string, args ...any)
}

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
func LoadTheme(configDir, themeName string, logger loggerInterface) *Theme {
	themeFolder := filepath.Join(configDir, "themes")
	if !utils.IsFolderExists(themeFolder) {
		logger.Debug("[MAIN] Themes folder not found. Extract themes to %q", themeFolder)
		err := extractThemeFiles(themeFolder)
		if err != nil {
			// We cannot extract themes. There is nothing we can do in this case
			// apart from applying the default theme and return.
			logger.Error("[MAIN] Cannot extract theme files: %v. Fallback to default theme.", err)
			theme := DefaultTheme()
			SetTheme(theme)
			return theme
		}
	}

	themeFile := filepath.Join(themeFolder, themeName+".json")
	logger.Info("[MAIN] Load theme from %q", themeFile)

	theme, err := loadThemeFromFile(themeFile)
	if err != nil {
		logger.Error("[MAIN] Cannot load theme: %v. Fallback to default theme.", err)
		theme = DefaultTheme()
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
	err = saveThemeToFile(data, filepath.Join(themesPath, "default.json"))
	if err != nil {
		return fmt.Errorf("failed to save default theme: %w", err)
	}

	// 2. Extract embedded themes
	entries, err := resources.Themes.ReadDir("themes")
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		// Cannot use filepath.Join in embedded filesystem, because there will be problems
		// with folder separators on Windows: "/" vs "\". Using path.Join instead.
		embeddedFSPath := path.Join("themes", entry.Name())
		data, err = resources.Themes.ReadFile(embeddedFSPath)
		if err != nil {
			return fmt.Errorf("failed to read embedded theme file: %w", err)
		}

		err = saveThemeToFile(data, filepath.Join(themesPath, entry.Name()))
		if err != nil {
			return fmt.Errorf("failed to save embedded theme file: %w", err)
		}
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

// ListAvailableThemes returns a list of available theme names.
func ListAvailableThemes(configDir string) ([]string, error) {
	themeFolder := filepath.Join(configDir, "themes")
	if !utils.IsFolderExists(themeFolder) {
		// logger.Debug("[MAIN] Themes folder not found. Extract themes to %q", themeFolder)
		err := extractThemeFiles(themeFolder)
		if err != nil {
			return nil, err
		}
	}

	entries, err := os.ReadDir(themeFolder)
	if err != nil {
		return nil, fmt.Errorf("failed to read themes directory: %w", err)
	}

	themes := []string{}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		themeName := strings.TrimSuffix(entry.Name(), ".json")
		themes = append(themes, themeName)
	}

	return themes, nil
}
