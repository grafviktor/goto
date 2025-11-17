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
	Warn(format string, args ...any)
	Error(format string, args ...any)
	Debug(format string, args ...any)
}

var currentTheme *Theme

// Set sets the current application theme.
func Set(theme *Theme) {
	currentTheme = theme
}

// Get returns the current application theme.
func Get() *Theme {
	if currentTheme == nil {
		currentTheme = DefaultTheme()
	}
	return currentTheme
}

// Load loads a theme from file or falls back to default.
func Load(configDir, themeName string, logger loggerInterface) error {
	themeFolder := filepath.Join(configDir, "themes")
	if !utils.IsFolderExists(themeFolder) {
		logger.Debug("[THEME] Themes folder not found. Extract themes to %q", themeFolder)
		extractThemeFiles(themeFolder, logger)
	}

	if utils.StringEmpty(&themeName) {
		logger.Warn("[THEME] Themes not set. Fallback to default theme: %q", DefaultTheme().Name)
		themeName = DefaultTheme().Name
	}

	themeFile := filepath.Join(themeFolder, themeName+".json")
	logger.Info("[THEME] Load theme from %q", themeFile)
	theme, err := loadThemeFromFile(themeFile)
	if err != nil {
		theme = DefaultTheme()
	}

	Set(theme)
	return err
}

// loadThemeFromFile loads a theme from a JSON file.
func loadThemeFromFile(filePath string) (*Theme, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var theme Theme
	if err = json.Unmarshal(data, &theme); err != nil {
		return nil, fmt.Errorf("failed to parse theme file: %w", err)
	}

	theme.computeStyles()
	return &theme, nil
}

func extractThemeFiles(themesPath string, logger loggerInterface) {
	// 1. Write default theme to disk
	data, err := json.MarshalIndent(DefaultTheme(), "", "  ")
	if err != nil {
		logger.Error("[THEME] Failed to marshal default theme: %v", err)
	}

	err = saveThemeToFile(data, filepath.Join(themesPath, "default.json"))
	if err != nil {
		logger.Error("[THEME] Failed to save default theme to %q: %v", filepath.Join(themesPath, "default.json"), err)
	}

	// 2. Extract embedded themes
	entries, err := resources.Themes.ReadDir("themes")
	if err != nil {
		logger.Error("[THEME] Failed to read themes directory: %v", err)

	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			logger.Debug("[THEME] Skip non-theme file in embedded filesystem: %q", entry.Name())
			continue
		}

		// Cannot use filepath.Join in embedded filesystem, because there will be problems
		// with folder separators on Windows: "/" vs "\". Using path.Join instead.
		embeddedFSPath := path.Join("themes", entry.Name())
		data, err = resources.Themes.ReadFile(embeddedFSPath)
		if err != nil {
			logger.Error("[THEME] Failed to read embedded theme file: %v", err)
		}

		err = saveThemeToFile(data, filepath.Join(themesPath, entry.Name()))
		if err != nil {
			logger.Error("[THEME] Failed to save embedded theme file to %q: %v", filepath.Join(themesPath, entry.Name()), err)
		}
	}
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

// ListInstalled returns a list of available theme names.
func ListInstalled(configDir string, logger loggerInterface) []string {
	themeFolder := filepath.Join(configDir, "themes")
	if !utils.IsFolderExists(themeFolder) {
		logger.Debug("[THEME] Themes folder not found. Extract themes to %q", themeFolder)
		extractThemeFiles(themeFolder, logger)
	}

	entries, err := os.ReadDir(themeFolder)
	if err != nil {
		logger.Error("[THEME] Failed to read themes directory: %v", err)
		return []string{DefaultTheme().Name}
	}

	themes := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		themeName := strings.TrimSuffix(entry.Name(), ".json")
		themes = append(themes, themeName)
	}

	return themes
}
