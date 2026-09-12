package config

import (
	"errors"
	"fmt"
	"strings"
)

// SupportedFeatures contains a list of application features that can be enabled or disabled.
var SupportedFeatures = []FeatureFlag{"ssh_config", "embedded_terminal"}

// FeatureFlag represents application feature flag that can be enabled or disabled.
type FeatureFlag string

func (ff *FeatureFlag) String() string {
	return string(*ff)
}

// Set validates and sets the feature flag value.
func (ff *FeatureFlag) Set(value string) error {
	var supportedFeatures []string
	for _, supported := range SupportedFeatures {
		if value == supported.String() {
			*ff = FeatureFlag(value)
			return nil
		}

		supportedFeatures = append(supportedFeatures, supported.String())
	}

	errMsg := fmt.Sprintf("\nsupported values: %s", strings.Join(supportedFeatures, "|"))
	return errors.New(errMsg)
}
