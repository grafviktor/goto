// Package state is in charge of storing and reading application state.
package state

import (
	"context"
	"fmt"
	"os"
	"path"
	"sync"

	"gopkg.in/yaml.v2"

	"github.com/grafviktor/goto/internal/config"
	"github.com/grafviktor/goto/internal/constant"
	"github.com/grafviktor/goto/internal/ui/theme"
	"github.com/grafviktor/goto/internal/utils"
	"github.com/grafviktor/goto/internal/version"
	"github.com/samber/lo"
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
		// 2. If default path is not exist, we must ingore this error(replicate what ssh does)
		// 3. If custom path is provided, but not exist, we must close the app with error
		// Probably we should not be too smart here and always ignore the error.
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

func (as *State) readFromFile() {
	var loadedState struct {
		State
		// Little hack - we want null values to be distinguishable from
		// zero values especially for boolean parameters.
		Theme            *string `yaml:"theme"`
		ScreenLayout     *string `yaml:"screen_layout"`
		SSHConfigEnabled *bool   `yaml:"enable_ssh_config"`
	}

	appStateFilePath := path.Join(as.AppHome, stateFile)
	as.Logger.Debug("[APPSTATE] Read application state from: %q", appStateFilePath)
	fileData, err := os.ReadFile(appStateFilePath)
	if err != nil {
		as.Logger.Warn("[APPSTATE] Can't read application state from file. Reason: %v", err)
	}

	err = yaml.Unmarshal(fileData, &loadedState)
	if err != nil {
		as.Logger.Error("[APPSTATE] Can't parse application state loaded from file. Reason: %v", err)
	}

	as.Group = loadedState.Group
	as.Selected = loadedState.Selected

	if loadedState.Theme == nil {
		as.Theme = theme.DefaultTheme().Name
	} else {
		as.Theme = *loadedState.Theme
	}

	if loadedState.ScreenLayout == nil {
		as.ScreenLayout = constant.ScreenLayoutDescription
	} else {
		as.ScreenLayout = constant.ScreenLayout(*loadedState.ScreenLayout)
	}

	if loadedState.SSHConfigEnabled == nil {
		// If there is no value for ssh config option, then we enable it by default
		as.SSHConfigEnabled = true
	} else {
		as.SSHConfigEnabled = *loadedState.SSHConfigEnabled
	}

	as.Logger.Debug("[APPSTATE] Screen layout: '%v'. Focused host id: '%v'", as.ScreenLayout, as.Selected)
}

func (st *State) applyConfig(cfg *config.Configuration) error {
	if !utils.StringEmpty(&cfg.LogLevel) {
		st.LogLevel = cfg.LogLevel
	}

	if cfg.DisableFeature != "" {
		// if disabled feature not equal to ssh config, then ssh config remains enabled
		st.SSHConfigEnabled = cfg.DisableFeature != config.FeatureSSHConfig
	}

	if cfg.EnableFeature != "" {
		// if enabled feature equal to ssh config, then ssh config becomes enabled
		st.SSHConfigEnabled = cfg.EnableFeature == config.FeatureSSHConfig
	}

	if !utils.StringEmpty(&cfg.SSHConfigFilePath) {
		userDefinedPath, err := utils.SSHConfigFilePath(cfg.SSHConfigFilePath)
		if err != nil {
			return fmt.Errorf("cannot set ssh config file path: %w", err)
		}
		st.SSHConfigFilePath = userDefinedPath
		st.IsUserDefinedSSHConfigPath = true
	}

	if !utils.StringEmpty(&cfg.SetTheme) {
		installedThemes := theme.ListInstalled(cfg.AppHome, st.Logger)
		if !lo.Contains(installedThemes, cfg.SetTheme) {
			return fmt.Errorf("theme %q is not available, installed themes: %v", cfg.SetTheme, installedThemes)
		}
		st.Theme = cfg.SetTheme
	}

	return nil
}

// Persist saves app state to disk.
func (as *State) Persist() error {
	appStateFilePath := path.Join(as.AppHome, stateFile)
	as.Logger.Debug("[APPSTATE] Persist application state to file: %q", appStateFilePath)
	result, err := yaml.Marshal(as)
	if err != nil {
		as.Logger.Error("[APPSTATE] Cannot marshall application state. %v", err)
		return err
	}

	err = os.WriteFile(appStateFilePath, result, 0o600)
	if err != nil {
		as.Logger.Error("[APPSTATE] Cannot save application state. %v", err)
		return err
	}

	return nil
}

// PrintConfig outputs user-definable parameters in the console.
func (as *State) PrintConfig() {
	version.Print()
	fmt.Println()
	fmt.Printf("App home:           %s\n", as.AppHome)
	fmt.Printf("Log level:          %s\n", as.LogLevel)
	fmt.Printf("SSH config enabled: %t\n", as.SSHConfigEnabled)
	if as.SSHConfigEnabled {
		fmt.Printf("SSH config path:    %s\n", as.SSHConfigFilePath)
	}
}

func (as *State) LogDetails(logger loggerInterface) {
	logger.Info("[CONFIG] Set application home folder to %q\n", as.AppHome)
	logger.Info("[CONFIG] Set application log level to %q\n", as.LogLevel)
	logger.Info("[CONFIG] SSH config enabled: %t\n", as.SSHConfigEnabled)
	logger.Info("[CONFIG] Set SSH config path to %q\n", as.SSHConfigFilePath)
}
