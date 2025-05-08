// Package config contains application configuration
package config

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

type iLogger interface {
	Debug(format string, args ...any)
	Info(format string, args ...any)
	Error(format string, args ...any)
	Close()
}

var supportedFeatures = []string{"ssh_config"}

type EnableFeature string

func (e *EnableFeature) String() string {
	return string(*e)
}

func (e *EnableFeature) Set(value string) error {
	for _, supported := range supportedFeatures {
		if value == supported {
			*e = EnableFeature(value)
			return nil
		}
	}

	errMsg := fmt.Sprintf("\nsupported values: %s", strings.Join(supportedFeatures, ", "))
	return errors.New(errMsg)
}

// User structs contains user-definable parameters.
type User struct {
	AppHome       string `env:"GG_HOME"`
	LogLevel      string `env:"GG_LOG_LEVEL" envDefault:"info"`
	SSHConfigFile string `env:"SSH_CONFIG_FILE"`
	EnableFeature EnableFeature
}

// Print outputs user-definable parameters in the console.
func (userConfig User) Print() {
	fmt.Printf("App home:   %s\n", userConfig.AppHome)
	fmt.Printf("Log level:  %s\n", userConfig.LogLevel)
	fmt.Printf("SSH config: %s\n", userConfig.SSHConfigFile)
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

	if len(cmdParams.SSHConfigFile) > 0 {
		envParams.SSHConfigFile = cmdParams.SSHConfigFile
	}
	logger.Debug("[CONFIG] Set SSH config path to '%s'\n", envParams.SSHConfigFile)

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
