// Package state is in charge of storing and reading application state.
//
//nolint:forbidigo // Use fmt.Printf for display user messages to stdout.
package state

import (
	"context"
	"fmt"
	"os"
	"path"
	"strings"
	"sync"

	"github.com/samber/lo"
	"gopkg.in/yaml.v2"

	"github.com/grafviktor/goto/internal/config"
	"github.com/grafviktor/goto/internal/constant"
	"github.com/grafviktor/goto/internal/ui/theme"
	"github.com/grafviktor/goto/internal/utils"
	"github.com/grafviktor/goto/internal/version"
)

type view int

const (
	// ViewHostList mode is active when we browse through a list of hostnames.
	ViewHostList view = iota
	// ViewGroupList mode is active when the app displays available host groups.
	ViewGroupList
	// ViewEditItem mode is active when we edit existing or add a new host.
	ViewEditItem
	// ViewMessage mode is active when there was an error when attempted to connect to a remote host.
	ViewMessage
)

// Once - this interface is used to avoid sync.Once restrictions in unit-tests.
type Once interface {
	Do(func())
}

var (
	st        *State
	once      Once = &sync.Once{}
	stateFile      = "state.yaml"
)

type loggerInterface interface {
	Debug(format string, args ...any)
	Warn(format string, args ...any)
	Info(format string, args ...any)
	Error(format string, args ...any)
	Close()
}

// State stores application state.
type State struct {
	AppHome                    string                `yaml:"-"`
	AppMode                    config.AppMode        `yaml:"-"`
	Context                    context.Context       `yaml:"-"`
	CurrentView                view                  `yaml:"-"`
	Group                      string                `yaml:"group,omitempty"`
	Height                     int                   `yaml:"-"`
	IsUserDefinedSSHConfigPath bool                  `yaml:"-"`
	Logger                     loggerInterface       `yaml:"-"`
	LogLevel                   string                `yaml:"-"`
	ScreenLayout               constant.ScreenLayout `yaml:"screen_layout,omitempty"`
	Selected                   int                   `yaml:"selected"`
	SSHConfigEnabled           bool                  `yaml:"enable_ssh_config"`
	SSHConfigFilePath          string                `yaml:"-"`
	Theme                      string                `yaml:"theme,omitempty"`
	Width                      int                   `yaml:"-"`
}

// Get - returns application state.
func Get() *State {
	return st
}

// IsInitialized - checks if the application state is initialized.
func IsInitialized() bool {
	return st != nil
}

// Initialize - creates application state.
func Initialize(ctx context.Context,
	cfg *config.Configuration,
	lg loggerInterface,
) (*State, error) {
	var err error

	once.Do(func() {
		lg.Debug("[APPSTATE] Create application state")

		// Cannot take this value from config because:
		// 1. We must distinguish between user-defined and default paths
		// 2. If default path does not exist, we must ignore this error(replicate what ssh does)
		// 3. If custom path is provided, but does not exist, we must close the app with error
		// Probably we should not be too smart here and always ignore the error or vice versa - always fail.
		var defaultSSHConfigPath string
		defaultSSHConfigPath, err = utils.SSHConfigDefaultFilePath()
		if err != nil {
			lg.Warn("[APPSTATE] Cannot determine default SSH config file path: %v.", err)
		}

		// Set default values
		st = &State{
			AppMode:           cfg.AppMode,
			AppHome:           cfg.AppHome,
			LogLevel:          cfg.LogLevel,
			Context:           ctx,
			Logger:            lg,
			SSHConfigFilePath: defaultSSHConfigPath,
		}

		// Read state from file
		st.readFromFile()
		// Apply configuration
		err = st.applyConfig(cfg)
	})

	return st, err
}

func (s *State) readFromFile() {
	var loadedState struct {
		Selected int    `yaml:"selected"`
		Group    string `yaml:"group"`
		// Little hack - we want to distinguish null values from
		// zero values especially for boolean parameters. Using pointers for that.
		Theme            *string `yaml:"theme"`
		ScreenLayout     *string `yaml:"screen_layout"`
		SSHConfigEnabled *bool   `yaml:"enable_ssh_config"`
	}

	appStateFilePath := path.Join(s.AppHome, stateFile)
	s.Logger.Debug("[APPSTATE] Read application state from: %q", appStateFilePath)
	fileData, err := os.ReadFile(appStateFilePath)
	if err != nil {
		s.Logger.Warn("[APPSTATE] Can't read application state from file. Reason: %v", err)
	}

	err = yaml.Unmarshal(fileData, &loadedState)
	if err != nil {
		s.Logger.Error("[APPSTATE] Can't parse application state loaded from file. Reason: %v", err)
	}

	s.Group = loadedState.Group
	s.Selected = loadedState.Selected

	if loadedState.Theme == nil {
		s.Theme = theme.DefaultTheme().Name
	} else {
		s.Theme = *loadedState.Theme
	}

	if loadedState.ScreenLayout == nil {
		s.ScreenLayout = constant.ScreenLayoutDescription
	} else {
		s.ScreenLayout = constant.ScreenLayout(*loadedState.ScreenLayout)
	}

	if loadedState.SSHConfigEnabled == nil {
		// If there is no value for ssh config option, then we enable it by default
		s.SSHConfigEnabled = true
	} else {
		s.SSHConfigEnabled = *loadedState.SSHConfigEnabled
	}

	s.Logger.Debug("[APPSTATE] Screen layout: '%v'. Focused host id: '%v'", s.ScreenLayout, s.Selected)
}

func (s *State) applyConfig(cfg *config.Configuration) error {
	if !utils.StringEmpty(&cfg.LogLevel) {
		s.LogLevel = cfg.LogLevel
	}

	if cfg.DisableFeature != "" {
		if cfg.DisableFeature == config.FeatureSSHConfig {
			s.SSHConfigEnabled = false
		} else {
			return fmt.Errorf("feature %q is not supported", cfg.DisableFeature)
		}
	}

	if cfg.EnableFeature != "" {
		if cfg.EnableFeature == config.FeatureSSHConfig {
			s.SSHConfigEnabled = true
		} else {
			return fmt.Errorf("feature %q is not supported", cfg.EnableFeature)
		}
	}

	if !utils.StringEmpty(&cfg.SSHConfigFilePath) {
		userDefinedPath, err := utils.SSHConfigFilePath(cfg.SSHConfigFilePath)
		if err != nil {
			return fmt.Errorf("cannot set ssh config file path: %w", err)
		}
		s.SSHConfigFilePath = userDefinedPath
		s.IsUserDefinedSSHConfigPath = true
	}

	if !utils.StringEmpty(&cfg.SetTheme) {
		installedThemes := theme.ListInstalled(cfg.AppHome, s.Logger)
		if !lo.Contains(installedThemes, cfg.SetTheme) {
			installedThemesStr := strings.Join(installedThemes, ", ")
			return fmt.Errorf("cannot find theme %q, installed themes: %v", cfg.SetTheme, installedThemesStr)
		}
		s.Theme = cfg.SetTheme
	}

	return nil
}

// Persist saves app state to disk.
func (s *State) Persist() error {
	appStateFilePath := path.Join(s.AppHome, stateFile)
	s.Logger.Debug("[APPSTATE] Persist application state to file: %q", appStateFilePath)
	result, err := yaml.Marshal(s)
	if err != nil {
		s.Logger.Error("[APPSTATE] Cannot marshall application state. %v", err)
		return err
	}

	err = os.WriteFile(appStateFilePath, result, 0o600)
	if err != nil {
		s.Logger.Error("[APPSTATE] Cannot save application state. %v", err)
		return err
	}

	return nil
}

func (s *State) print() {
	fmt.Printf("App home:          %s\n", s.AppHome)
	fmt.Printf("Log level:         %s\n", s.LogLevel)
	fmt.Printf("SSH config status: %s\n", lo.Ternary(s.SSHConfigEnabled, "enabled", "disabled"))
	if s.SSHConfigEnabled {
		fmt.Printf("SSH config path:   %s\n", s.SSHConfigFilePath)
	}
}

// PrintConfig outputs user-definable parameters in the console.
func (s *State) PrintConfig() {
	version.Print()
	fmt.Println()
	s.print()
}

func (s *State) LogDetails(logger loggerInterface) {
	logger.Info("[CONFIG] Application home folder: %q\n", s.AppHome)
	logger.Info("[CONFIG] Application log level:   %q\n", s.LogLevel)
	logger.Info("[CONFIG] SSH config status:       %q\n", lo.Ternary(s.SSHConfigEnabled, "enabled", "disabled"))
	if s.SSHConfigEnabled {
		logger.Info("[CONFIG] SSH config path:         %q\n", s.SSHConfigFilePath)
	}
}
