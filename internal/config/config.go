package config

import (
	"context"
)

// const configFileName = "config.yaml"

type Logger interface {
	Debug(format string, args ...any)
	Error(format string, args ...any)
	Close()
}

type User struct {
	AppHome  string `yaml:"-" env:"GG_HOME"`
	LogLevel string `yaml:"log_level,omitempty" env:"GG_LOG_LEVEL" envDefault:"info"`
}

func Merge(envParams, cmdParams User, logger Logger) User {
	// TODO: Can use reflection to override envParams with cmdParams, instead of addressing exact fields
	if len(cmdParams.AppHome) > 0 {
		envParams.AppHome = cmdParams.AppHome
	}
	logger.Debug("Set application home folder to %s\n", envParams.AppHome)

	if len(cmdParams.LogLevel) > 0 {
		envParams.LogLevel = cmdParams.LogLevel
	}
	logger.Debug("Set application log level to %s\n", envParams.LogLevel)

	return envParams
}

type Application struct {
	Context context.Context
	Logger  Logger
	Config  User
}

func NewApplication(ctx context.Context, userConfig User, logger Logger) Application {
	app := Application{
		Context: ctx,
		Config:  userConfig,
		Logger:  logger,
	}

	return app
}
