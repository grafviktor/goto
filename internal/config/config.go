package config

import (
	"context"
	"os"
	"path"

	"github.com/grafviktor/goto/internal/utils"
	"gopkg.in/yaml.v2"
)

var appName = "goto"
var configFileName = "config.yaml"

type Logger interface {
	Debug(format string, args ...any)
	Close()
}
type UserSettings struct {
	logger         Logger
	configFilePath string
	homeFolder     string
	HostsFilePath  string `env:"HOSTS_FILE" yaml:"host_file,omitempty"`
	LogLevel       string `env:"LOG_LEVEL"  yaml:"log_level,omitempty"`
	LogFilePath    string `env:"LOG_PATH"   yaml:"log_path, omitempty"`
}

func NewUserSettings(logger Logger) (UserSettings, error) {
	us := UserSettings{logger: logger}

	appHome, err := utils.GetAppDir(logger, appName)
	if err != nil {
		logger.Debug("Failed to get application home folder%v\n", err)

		return us, err
	}

	configFilePath := path.Join(appHome, configFileName)

	logger.Debug("Read application configuration from %s\n", configFilePath)
	fileData, err := os.ReadFile(configFilePath)
	if err != nil {
		logger.Debug("Can't read application configuration %v\n", err)

	}

	us.homeFolder = appHome

	err = yaml.Unmarshal(fileData, &us)
	if err != nil {
		us.logger.Debug("Can't read parse application configuration %v\n", err)

		return us, err
	}

	return us, nil

}

func (us *UserSettings) Save() error {
	result, err := yaml.Marshal(us)
	if err != nil {
		return err
	}

	err = os.WriteFile(us.configFilePath, result, 0o600)
	if err != nil {
		return err
	}

	return nil
}

func New(ctx context.Context, logger Logger, settings UserSettings) Application {

	config := Application{
		Context:      ctx,
		Logger:       logger,
		AppName:      appName,
		UserSettings: settings,
	}

	return config
}

type Application struct {
	HomeFolder string
	AppName    string
	Context    context.Context
	Logger     Logger
	UserSettings
}
