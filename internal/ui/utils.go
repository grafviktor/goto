package ui

import (
	"os"
	"path"

	"github.com/grafviktor/goto/internal/config"
	"github.com/grafviktor/goto/internal/model"
	"gopkg.in/yaml.v2"
)

func loadUIState(config config.Application) (model.AppState, error) {
	statePath := path.Join(config.AppHome, "state.yaml")
	appState := model.AppState{}

	config.Logger.Debug("Read application state from %s\n", statePath)
	fileData, err := os.ReadFile(statePath)
	if err != nil {
		config.Logger.Debug("Can't read application state %v\n", err)
		return appState, err
	}

	err = yaml.Unmarshal(fileData, &appState)
	if err != nil {
		config.Logger.Debug("Can't read parse application state %v\n", err)
		return appState, err
	}

	return appState, nil
}

func saveUIState(appState model.AppState) error {
	/*result, err := yaml.Marshal(appState)
	if err != nil {
		return err
	}

	// FIXME: configPath is not here
	err = os.WriteFile(app.configPath, result, 0o600)
	if err != nil {
		return err
	}*/

	return nil
}
