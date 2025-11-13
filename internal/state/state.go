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
	Selected         int `yaml:"selected"`
	appStateFilePath string
	Logger           loggerInterface       `yaml:"-"`
	CurrentView      view                  `yaml:"-"`
	Width            int                   `yaml:"-"`
	Height           int                   `yaml:"-"`
	ScreenLayout     constant.ScreenLayout `yaml:"screenLayout,omitempty"`
	Group            string                `yaml:"group,omitempty"`
	// SSHConfigEnabled is a part of ApplicationState, not user config, because it is a feature flag
	// which is persisted across application restarts. In other words, once defined, it will be
	// persisted in the state.yaml file and will be used in the next application run.
	SSHConfigEnabled bool   `yaml:"enable_ssh_config"`
	Theme            string `yaml:"theme,omitempty"`
	// ApplicationConfig *config.Configuration `yaml:"-"`
	Context                    context.Context `yaml:"-"`
	AppHome                    string
	LogLevel                   string
	SSHConfigFilePath          string
	AppMode                    config.AppMode
	IsUserDefinedSSHConfigPath bool
}

// Initialize - creates application state.
func Initialize(ctx context.Context,
	cfg *config.Configuration,
	lg loggerInterface,
) (*State, error) {
	var err error

	once.Do(func() {
		lg.Debug("[APPSTATE] Create application state")
		st = &State{
			AppMode:          cfg.AppMode,
			AppHome:          cfg.AppHome,
			appStateFilePath: path.Join(cfg.AppHome, stateFile),
			Context:          ctx,
			Logger:           lg,
			LogLevel:         cfg.LogLevel,
			SSHConfigEnabled: true,
		}

		err = st.readFromFile()
		applyConfig(st, cfg)
	})

	return st, err
}

func applyConfig(st *State, cfg *config.Configuration) {
	if cfg.DisableFeature != "" {
		// if disabled feature not equal to ssh config, then ssh config remains enabled
		st.SSHConfigEnabled = cfg.DisableFeature != config.FeatureSSHConfig
	}

	if cfg.EnableFeature != "" {
		// if enabled feature equal to ssh config, then ssh config becomes enabled
		st.SSHConfigEnabled = cfg.EnableFeature == config.FeatureSSHConfig
	}

	if !utils.StringEmpty(&cfg.SSHConfigFilePath) {
		st.IsUserDefinedSSHConfigPath = true
	}

	if !utils.StringEmpty(&cfg.SetTheme) {
		st.Theme = cfg.SetTheme
	}

	if !utils.StringEmpty(&cfg.SSHConfigFilePath) {
		st.SSHConfigFilePath = cfg.SSHConfigFilePath
	}
}

// Get - returns application state.
func Get() *State {
	return st
}

// IsInitialized - checks if the application state is initialized.
func IsInitialized() bool {
	return st != nil
}

func (as *State) readFromFile() error {
	as.Logger.Debug("[APPSTATE] Read application state from: %q", as.appStateFilePath)
	fileData, err := os.ReadFile(as.appStateFilePath)
	if err != nil {
		as.Logger.Info("[APPSTATE] Can't read application state from file '%v'", err)
		return err
	}

	err = yaml.Unmarshal(fileData, as)
	if err != nil {
		as.Logger.Error("[APPSTATE] Can't parse application state loaded from file '%v'", err)
		return err
	}

	as.Logger.Debug("[APPSTATE] Screen layout: '%v'. Focused host id: '%v'", as.ScreenLayout, as.Selected)

	return nil
}

// Persist saves app state to disk.
func (as *State) Persist() error {
	as.Logger.Debug("[APPSTATE] Persist application state to file: %q", as.appStateFilePath)
	result, err := yaml.Marshal(as)
	if err != nil {
		as.Logger.Error("[APPSTATE] Cannot marshall application state. %v", err)
		return err
	}

	err = os.WriteFile(as.appStateFilePath, result, 0o600)
	if err != nil {
		as.Logger.Error("[APPSTATE] Cannot save application state. %v", err)
		return err
	}

	return nil
}

func (as *State) printConfig() {
	fmt.Printf("App home:           %s\n", as.AppHome)
	fmt.Printf("Log level:          %s\n", as.LogLevel)
	if as.SSHConfigEnabled {
		fmt.Printf("SSH config enabled: %t\n", as.SSHConfigEnabled)
		fmt.Printf("SSH config path:    %s\n", as.SSHConfigFilePath)
	}
}

// PrintConfig outputs user-definable parameters in the console.
func (as *State) PrintConfig() {
	version.Print()
	fmt.Println()
	as.printConfig()
}

func (as *State) LogDetails(logger loggerInterface) {
	logger.Info("[CONFIG] Set application home folder to %q\n", as.AppHome)
	logger.Info("[CONFIG] Set application log level to %q\n", as.LogLevel)
	logger.Info("[CONFIG] SSH config enabled: %t\n", as.SSHConfigEnabled)
	logger.Info("[CONFIG] Set SSH config path to %q\n", as.SSHConfigFilePath)
}
