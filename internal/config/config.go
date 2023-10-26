package config

import (
	"context"
	"os"
	"path"

	"github.com/grafviktor/goto/internal/model"
	"github.com/grafviktor/goto/internal/utils"
	"gopkg.in/yaml.v2"
)

var appName = "goto"
var configFile = "config.yaml"

type Logger interface {
	Debug(format string, args ...any)
	Close()
}

func New(ctx context.Context, logger Logger) Application {
	// TODO:
	// AppHome should come from main module and can be overwritten by a user
	// AppHome cannot be read from a config file
	appHome, err := utils.GetAppDir(logger, appName)
	if err != nil {
		logger.Debug("Could not get application home folder")
	}

	config := Application{
		Context: ctx,
		Logger:  logger,
		AppHome: appHome,
		AppName: appName,
	}

	err = config.load()
	if err != nil {
		logger.Debug("Could not load application from a config file")
	}

	return config
}

type Application struct {
	AppHome    string
	AppName    string
	Context    context.Context
	Logger     Logger
	configPath string
	model.AppConfig
}

func (app *Application) load() error {
	var appConfigModel model.AppConfig
	app.configPath = path.Join(app.AppHome, configFile)

	app.Logger.Debug("Read application configuration from %s\n", app.configPath)
	fileData, err := os.ReadFile(app.configPath)
	if err != nil {
		app.Logger.Debug("Can't read application configuration %v\n", err)
		return err
	}

	err = yaml.Unmarshal(fileData, &appConfigModel)
	if err != nil {
		app.Logger.Debug("Can't read parse application configuration %v\n", err)
		return err
	}

	app.AppConfig = appConfigModel

	return nil
}

func (app *Application) Save() error {
	result, err := yaml.Marshal(app.AppConfig)
	if err != nil {
		return err
	}

	err = os.WriteFile(app.configPath, result, 0o600)
	if err != nil {
		return err
	}

	return nil
}
