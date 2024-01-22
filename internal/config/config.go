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
	AppHome  string `env:"GG_HOME"`
	LogLevel string `env:"GG_LOG_LEVEL" envDefault:"info"`
}

// Print outputs user-definable parameters in the console.
func (userConfig User) Print() {
	fmt.Printf("App home:  %s\n", userConfig.AppHome)
	fmt.Printf("Log level: %s\n", userConfig.LogLevel)
}

// Merge builds application configuration from user parameters and common objects. For instance - logger.
func Merge(envParams, cmdParams User, logger iLogger) User {
	//nolint:godox
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
