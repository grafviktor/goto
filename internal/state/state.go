// Package state is in charge of storing and reading application state.
package state

import (
	"fmt"
	"os"
	"path"
	"sync"

	"gopkg.in/yaml.v2"

	"github.com/grafviktor/goto/internal/config"
	"github.com/grafviktor/goto/internal/constant"
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

var (
	appState  *Application
	once      sync.Once
	stateFile = "state.yaml"
)

type iLogger interface {
	Debug(format string, args ...any)
	Info(format string, args ...any)
	Error(format string, args ...any)
}

// Application stores application state.
type Application struct {
	Selected         int `yaml:"selected"`
	appStateFilePath string
	logger           iLogger
	CurrentView      view                  `yaml:"-"`
	Width            int                   `yaml:"-"`
	Height           int                   `yaml:"-"`
	ScreenLayout     constant.ScreenLayout `yaml:"screenLayout,omitempty"`
	Group            string                `yaml:"group,omitempty"`
	// SSHConfigEnabled is a part of ApplicationState, not user config, because it is a feature flag
	// which is persisted across application restarts. In other words, once defined, it will be
	// persisted in the state.yaml file and will be used in the next application run.
	SSHConfigEnabled  bool               `yaml:"ssh_config"`
	ApplicationConfig config.Application `yaml:"-"`
}

// Create - creates application state.
func Create(appConfig config.Application, lg iLogger) *Application {
	once.Do(func() {
		lg.Debug("[APPSTATE] Create application state")
		appState = &Application{
			appStateFilePath:  path.Join(appConfig.UserConfig.AppHome, stateFile),
			logger:            lg,
			Group:             "",
			SSHConfigEnabled:  true,
			ApplicationConfig: appConfig,
		}

		// If we cannot read previously created application state, that's fine - we can continue execution.
		lg.Debug("[APPSTATE] Application state is not ready, restore from file")
		_ = appState.readFromFile()
	})

	return appState
}

// Get - returns application state.
func Get() *Application {
	return appState
}

func (as *Application) readFromFile() error {
	as.logger.Debug("[APPSTATE] Read application state from: '%s'", as.appStateFilePath)
	fileData, err := os.ReadFile(as.appStateFilePath)
	if err != nil {
		as.logger.Info("[APPSTATE] Can't read application state from file '%v'", err)
		return err
	}

	err = yaml.Unmarshal(fileData, as)
	if err != nil {
		as.logger.Error("[APPSTATE] Can't parse application state loaded from file '%v'", err)
		return err
	}

	as.logger.Debug("[APPSTATE] Screen layout: '%v'. Focused host id: '%v'", as.ScreenLayout, as.Selected)

	return nil
}

// Persist saves app state to disk.
func (as *Application) Persist() error {
	as.logger.Debug("[APPSTATE] Persist application state to file: %s", as.appStateFilePath)
	result, err := yaml.Marshal(as)
	if err != nil {
		as.logger.Error("[APPSTATE] Cannot marshall application state. %v", err)
		return err
	}

	err = os.WriteFile(as.appStateFilePath, result, 0o600)
	if err != nil {
		as.logger.Error("[APPSTATE] Cannot save application state. %v", err)
		return err
	}

	return nil
}

// Print outputs user-definable parameters in the console.
func (as *Application) PrintConfig() {
	userConfig := as.ApplicationConfig.UserConfig
	fmt.Printf("App home:           %s\n", userConfig.AppHome)
	fmt.Printf("Log level:          %s\n", userConfig.LogLevel)
	if as.SSHConfigEnabled {
		fmt.Printf("SSH config enabled: %t\n", as.SSHConfigEnabled)
		fmt.Printf("SSH config path:    %s\n", userConfig.SSHConfigFilePath)
	}
}
