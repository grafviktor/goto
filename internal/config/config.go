// Package config contains application configuration
package config

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/grafviktor/goto/internal/state"
)

type iLogger interface {
	Debug(format string, args ...any)
	Info(format string, args ...any)
	Error(format string, args ...any)
	Close()
}

var SupportedFeatures = []string{"ssh_config"}

type FeatureFlag string

func (ff *FeatureFlag) String() string {
	return string(*ff)
}

func (ff *FeatureFlag) Set(value string) error {
	for _, supported := range SupportedFeatures {
		if value == supported {
			*ff = FeatureFlag(value)
			return nil
		}
	}

	errMsg := fmt.Sprintf("\nsupported values: %s", strings.Join(SupportedFeatures, ", "))
	return errors.New(errMsg)
}

// User structs contains user-definable parameters.
type User struct {
	AppHome           string `env:"GG_HOME"`
	LogLevel          string `env:"GG_LOG_LEVEL" envDefault:"info"`
	SSHConfigFilePath string `env:"SSH_CONFIG_FILE_PATH"`
	EnableFeature     FeatureFlag
	DisableFeature    FeatureFlag
}

// Print outputs user-definable parameters in the console.
func (userConfig User) Print() {
	fmt.Printf("App home:           %s\n", userConfig.AppHome)
	fmt.Printf("Log level:          %s\n", userConfig.LogLevel)
	appState := state.Get()
	if appState.SSHConfigEnabled {
		fmt.Printf("SSH config enabled: %t\n", appState.SSHConfigEnabled)
		fmt.Printf("SSH config path:    %s\n", userConfig.SSHConfigFilePath)
	}
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

	if len(cmdParams.EnableFeature) > 0 {
		envParams.EnableFeature = cmdParams.EnableFeature
	}
	logger.Debug("[CONFIG] Enable feature '%s'\n", envParams.EnableFeature)

	if len(cmdParams.DisableFeature) > 0 {
		envParams.DisableFeature = cmdParams.DisableFeature
	}
	logger.Debug("[CONFIG] Disable feature '%s'\n", envParams.DisableFeature)

	if len(cmdParams.SSHConfigFilePath) > 0 {
		envParams.SSHConfigFilePath = cmdParams.SSHConfigFilePath
	}
	logger.Debug("[CONFIG] Set SSH config path to '%s'\n", envParams.SSHConfigFilePath)

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
