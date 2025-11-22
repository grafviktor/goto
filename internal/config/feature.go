package config

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
