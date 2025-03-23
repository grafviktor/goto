// Package config contains application configuration
package config

import (
	"context"
	"fmt"
)

type iLogger interface {
	Debug(format string, args ...any)
	Info(format string, args ...any)
	Error(format string, args ...any)
	Close()
}

// User structs contains user-definable parameters.
type User struct {
	AppHome   string `env:"GG_HOME,expand" envDefault:"${APP_DEFAULT_HOME}"` // expand is used to parse placeholders
	LogLevel  string `env:"GG_LOG_LEVEL" envDefault:"info"`
	SSHConfig string `env:"SSH_CONFIG,expand" envDefault:"${SSH_DEFAULT_CONFIG}"`
}

// Print outputs user-definable parameters in the console.
func (userConfig User) Print() {
	fmt.Printf("App home:   %s\n", userConfig.AppHome)
	fmt.Printf("Log level:  %s\n", userConfig.LogLevel)
	fmt.Printf("SSH config: %s\n", userConfig.SSHConfig)
}

// Merge builds application configuration from user parameters and common objects. For instance - logger.
func Merge(envParams, cmdParams User, logger iLogger) User {
	if len(cmdParams.AppHome) > 0 {
		envParams.AppHome = cmdParams.AppHome
	}
	logger.Debug("[CONFIG] Set application home folder to '%s'\n", envParams.AppHome)

	if len(cmdParams.LogLevel) > 0 {
		envParams.LogLevel = cmdParams.LogLevel
	}
	logger.Debug("[CONFIG] Set application log level to '%s'\n", envParams.LogLevel)

	if len(cmdParams.SSHConfig) > 0 {
		envParams.SSHConfig = cmdParams.SSHConfig
	}
	logger.Debug("[CONFIG] Set SSH config path to '%s'\n", envParams.SSHConfig)

	return envParams
}

// Application is a struct which contains logger, application context and user parameters.
type Application struct {
	Context context.Context
	Logger  iLogger
	Config  User
}

// NewApplication constructs application configuration.
func NewApplication(ctx context.Context, userConfig User, logger iLogger) Application {
	app := Application{
		Context: ctx,
		Config:  userConfig,
		Logger:  logger,
	}

	return app
}
