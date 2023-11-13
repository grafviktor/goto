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

type logger interface {
	Debug(format string, args ...any)
}

type ApplicationState struct {
	Selected         int `yaml:"selected"`
	appStateFilePath string
	logger           logger
	Width            int `yaml:"-"`
	Height           int `yaml:"-"`
}

func Get(appHomePath string, lg logger) *ApplicationState {
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
