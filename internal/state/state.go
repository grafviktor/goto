// Package state is in charge of storing and reading application state.
package state

import (
	"os"
	"path"
	"sync"

	"gopkg.in/yaml.v2"

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
	appState  *ApplicationState
	once      sync.Once
	stateFile = "state.yaml"
)

type iLogger interface {
	Debug(format string, args ...any)
	Info(format string, args ...any)
	Error(format string, args ...any)
}

// ApplicationState stores application state.
type ApplicationState struct {
	Selected         int `yaml:"selected"`
	appStateFilePath string
	logger           iLogger
	CurrentView      view                  `yaml:"-"`
	Width            int                   `yaml:"-"`
	Height           int                   `yaml:"-"`
	ScreenLayout     constant.ScreenLayout `yaml:"screenLayout,omitempty"`
	Group            string                `yaml:"group,omitempty"`
}

// Get - reads application state from disk.
func Get(appHomePath string, lg iLogger) *ApplicationState {
	lg.Debug("[APPSTATE] Get application state")
	once.Do(func() {
		appState = &ApplicationState{
			appStateFilePath: path.Join(appHomePath, stateFile),
			logger:           lg,
			Group:            "", // TODO: Find a better name for this field
		}

		// If we cannot read previously created application state, that's fine - we can continue execution.
		lg.Debug("[APPSTATE] Application state is not ready, restore from file")
		_ = appState.readFromFile()
	})

	return appState
}

func (as *ApplicationState) readFromFile() error {
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
func (as *ApplicationState) Persist() error {
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
