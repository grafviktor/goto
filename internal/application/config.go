// Package application config contains configuration struct
package application

import (
	"errors"
	"fmt"
	"strings"
)

// SupportedFeatures contains a list of application features that can be enabled or disabled.
var SupportedFeatures = []string{"ssh_config"}

// FeatureFlag represents application feature flag that can be enabled or disabled.
type FeatureFlag string

func (ff *FeatureFlag) String() string {
	return string(*ff)
}

// Set validates and sets the feature flag value.
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

// Configuration structs contains user-definable parameters.
type Configuration struct {
	AppHome                          string `env:"GG_HOME"`
	LogLevel                         string `env:"GG_LOG_LEVEL"            envDefault:"info"`
	SSHConfigFilePath                string `env:"GG_SSH_CONFIG_FILE_PATH"`
	IsSSHConfigFilePathDefinedByUser bool
	DisplayVersionAndExit            bool
	EnableFeature                    FeatureFlag
	DisableFeature                   FeatureFlag
	SetTheme                         string
}
