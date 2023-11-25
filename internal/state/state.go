// Package state is in charge of storing and reading application state.
package state

import (
	"os"
	"path"
	"sync"

	"gopkg.in/yaml.v2"
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
	Width            int `yaml:"-"`
	Height           int `yaml:"-"`
}

// Get - reads application stat from disk.
func Get(appHomePath string, lg iLogger) *ApplicationState {
	once.Do(func() {
		appState = &ApplicationState{
			appStateFilePath: path.Join(appHomePath, stateFile),
			logger:           lg,
		}

		// if we cannot read previously created application state, that's fine
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
