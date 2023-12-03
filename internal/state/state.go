// Package state is in charge of storing and reading application state.
package state

import (
	"os"
	"path"
	"sync"

	"gopkg.in/yaml.v2"
)

type view int

const (
	// ViewHostList mode is active when we browse through a list of hostnames.
	ViewHostList view = iota
	// ViewEditItem mode is active when we edit existing or add a new host.
	ViewEditItem
	// ViewErrorMessage mode is active when there was an error when attenpted to connect to a remote host.
	ViewErrorMessage
)

var (
	appState  *ApplicationState
	once      sync.Once
	stateFile = "state.yaml"
)

type iLogger interface {
	Debug(format string, args ...any)
}

// ApplicationState stores application state.
type ApplicationState struct {
	Selected         int `yaml:"selected"`
	appStateFilePath string
	logger           iLogger
	CurrentView      view  `yaml:"-"`
	Err              error `yaml:"-"`
	Width            int   `yaml:"-"`
	Height           int   `yaml:"-"`
}

// Get - reads application stat from disk.
func Get(appHomePath string, lg iLogger) *ApplicationState {
	once.Do(func() {
		appState = &ApplicationState{
			appStateFilePath: path.Join(appHomePath, stateFile),
			logger:           lg,
		}

		// If we cannot read previously created application state, that's fine - we can continue execution.
		_ = appState.readFromFile()
	})

	return appState
}

func (as *ApplicationState) readFromFile() error {
	as.logger.Debug("Read application state from %s\n", as.appStateFilePath)
	fileData, err := os.ReadFile(as.appStateFilePath)
	if err != nil {
		as.logger.Debug("Can't read application state %v\n", err)
		return err
	}

	err = yaml.Unmarshal(fileData, as)
	if err != nil {
		as.logger.Debug("Can't read parse application state %v\n", err)
		return err
	}

	return nil
}

// Persist saves app state to disk.
func (as *ApplicationState) Persist() error {
	result, err := yaml.Marshal(as)
	if err != nil {
		return err
	}

	err = os.WriteFile(as.appStateFilePath, result, 0o600)
	if err != nil {
		return err
	}

	return nil
}
