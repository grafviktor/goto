// Package state is in charge of storing and reading application state.
package state

import (
	"context"
	"fmt"
	"os"
	"path"
	"sync"

	"gopkg.in/yaml.v2"

	"github.com/grafviktor/goto/internal/application"
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
	Warn(format string, args ...any)
	Info(format string, args ...any)
	Error(format string, args ...any)
}

// Application stores application state.
type Application struct {
	Selected         int `yaml:"selected"`
	appStateFilePath string
	Logger           iLogger               `yaml:"-"`
	CurrentView      view                  `yaml:"-"`
	Width            int                   `yaml:"-"`
	Height           int                   `yaml:"-"`
	ScreenLayout     constant.ScreenLayout `yaml:"screenLayout,omitempty"`
	Group            string                `yaml:"group,omitempty"`
	// SSHConfigEnabled is a part of ApplicationState, not user config, because it is a feature flag
	// which is persisted across application restarts. In other words, once defined, it will be
	// persisted in the state.yaml file and will be used in the next application run.
	SSHConfigEnabled  bool                      `yaml:"ssh_config"`
	ApplicationConfig application.Configuration `yaml:"-"`
	Context           context.Context           `yaml:"-"`
}

// Create - creates application state.
func Create(appContext context.Context,
	appConfig application.Configuration,
	lg iLogger,
) *Application {
	once.Do(func() {
		lg.Debug("[APPSTATE] Create application state")
		appState = &Application{
			appStateFilePath:  path.Join(appConfig.AppHome, stateFile),
			Logger:            lg,
			Group:             "",
			SSHConfigEnabled:  true,
			ApplicationConfig: appConfig,
			Context:           appContext,
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

// IsInitialized - checks if the application state is initialized.
func IsInitialized() bool {
	return appState != nil
}

func (as *Application) readFromFile() error {
	as.Logger.Debug("[APPSTATE] Read application state from: '%s'", as.appStateFilePath)
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
func (as *Application) Persist() error {
	as.Logger.Debug("[APPSTATE] Persist application state to file: %s", as.appStateFilePath)
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

// Print outputs user-definable parameters in the console.
func (as *Application) PrintConfig() {
	fmt.Printf("App home:           %s\n", as.ApplicationConfig.AppHome)
	fmt.Printf("Log level:          %s\n", as.ApplicationConfig.LogLevel)
	if as.SSHConfigEnabled {
		fmt.Printf("SSH config enabled: %t\n", as.SSHConfigEnabled)
		fmt.Printf("SSH config path:    %s\n", as.ApplicationConfig.SSHConfigFilePath)
	}
}
