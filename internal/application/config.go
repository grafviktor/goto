// Package config contains application configuration
package application

import (
	"errors"
	"fmt"
	"strings"
)

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

// TODO: Rename from application.Configuration to config.Parameter

// Configuration structs contains user-definable parameters.
type Configuration struct {
	AppHome               string `env:"GG_HOME"`
	LogLevel              string `env:"GG_LOG_LEVEL" envDefault:"info"`
	SSHConfigFilePath     string `env:"GG_SSH_CONFIG_FILE_PATH"`
	DisplayVersionAndExit bool
	EnableFeature         FeatureFlag
	DisableFeature        FeatureFlag
}

// Merge builds application configuration from user parameters and common objects.
func Merge(envParams, cmdParams Configuration) Configuration {
	if len(cmdParams.AppHome) > 0 {
		envParams.AppHome = cmdParams.AppHome
	}

	if len(cmdParams.LogLevel) > 0 {
		envParams.LogLevel = cmdParams.LogLevel
	}

	if len(cmdParams.EnableFeature) > 0 {
		envParams.EnableFeature = cmdParams.EnableFeature
	}

	if len(cmdParams.DisableFeature) > 0 {
		envParams.DisableFeature = cmdParams.DisableFeature
	}

	if len(cmdParams.SSHConfigFilePath) > 0 {
		envParams.SSHConfigFilePath = cmdParams.SSHConfigFilePath
	}

	return envParams
}
