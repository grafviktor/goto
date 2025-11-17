// Package config contains application configuration struct
package config

import (
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/caarlos0/env/v10"
	"github.com/grafviktor/goto/internal/utils"
)

const (
	appName          = "goto"
	FeatureSSHConfig = "ssh_config"
)

// Configuration structs contains user-definable parameters.
type Configuration struct {
	AppMode               AppMode
	AppName               string
	DisableFeature        FeatureFlag
	DisplayVersionAndExit bool
	EnableFeature         FeatureFlag
	SetSSHConfigEnabled   bool
	SetTheme              string
	AppHome               string `env:"GG_HOME"`
	LogLevel              string `env:"GG_LOG_LEVEL"            envDefault:"info"`
	SSHConfigFilePath     string `env:"GG_SSH_CONFIG_FILE_PATH"`
}

func Initialize() (*Configuration, error) {
	envConfig, err := parseEnvironmentVariables()
	if err != nil {
		return envConfig, fmt.Errorf("failed to parse environment variables: %w", err)
	}

	cmdConfig, err := parseCommandLineFlags(envConfig)
	if err != nil {
		return envConfig, err
	}

	appConfig, err := setConfigDefaults(cmdConfig)
	if err != nil {
		return envConfig, fmt.Errorf("failed to set configuration defaults: %w", err)
	}

	return appConfig, nil
}

// parseEnvironmentConfig parses environment configuration.
func parseEnvironmentVariables() (*Configuration, error) {
	envConfig := &Configuration{}
	err := env.Parse(envConfig)
	if err != nil {
		return envConfig, fmt.Errorf("error parsing environment configuration: %w", err)
	}

	return envConfig, nil
}

// parseCommandLineFlags parses command line flags and returns the configuration.
func parseCommandLineFlags(envConfig *Configuration) (*Configuration, error) {
	cmdConfig := &Configuration{AppMode: AppModeType.StartUI}

	// Command line parameters have the highest precedence
	flag.BoolVar(&cmdConfig.DisplayVersionAndExit, "v", false, "Display application details")
	flag.StringVar(&cmdConfig.AppHome, "f", envConfig.AppHome, "Application home folder")
	flag.StringVar(&cmdConfig.LogLevel, "l", envConfig.LogLevel, "Log verbosity level: debug, info")
	flag.StringVar(
		&cmdConfig.SSHConfigFilePath,
		"s",
		envConfig.SSHConfigFilePath,
		"Specifies an alternative per-user SSH configuration file path",
	)
	flag.Var(
		&cmdConfig.EnableFeature,
		"e",
		fmt.Sprintf("Enable feature. Supported values: %s", strings.Join(SupportedFeatures, "|")),
	)
	flag.Var(
		&cmdConfig.DisableFeature,
		"d",
		fmt.Sprintf("Disable feature. Supported values: %s", strings.Join(SupportedFeatures, "|")),
	)
	flag.StringVar(&cmdConfig.SetTheme, "set-theme", "", "Set application theme")
	flag.Parse()

	var err error
	switch {
	case cmdConfig.DisplayVersionAndExit:
		handleDisplayVersion(cmdConfig)
	case cmdConfig.EnableFeature != "":
		fmt.Printf("[CONFIG] Enable feature %q\n", cmdConfig.EnableFeature.String())
		handleFeatureToggle(cmdConfig, cmdConfig.EnableFeature.String(), true)
	case cmdConfig.DisableFeature != "":
		fmt.Printf("[CONFIG] Disable feature %q\n", cmdConfig.DisableFeature.String())
		handleFeatureToggle(cmdConfig, cmdConfig.DisableFeature.String(), false)
	case cmdConfig.SetTheme != "":
		fmt.Printf("[CONFIG] Set theme to %q\n", cmdConfig.SetTheme)
		cmdConfig.AppMode = AppModeType.HandleParam
	}

	return cmdConfig, err
}

func setConfigDefaults(config *Configuration) (*Configuration, error) {
	var err error
	config.AppName = appName
	config.AppHome, err = utils.AppDir(appName, config.AppHome)
	if err != nil {
		log.Printf("[CONFIG] Application home folder error: %v", err)
	}

	return config, nil
}

func handleDisplayVersion(config *Configuration) {
	config.AppMode = AppModeType.DisplayInfo
}

// handleFeatureToggle handles enabling or disabling features.
func handleFeatureToggle(config *Configuration, featureName string, enable bool) {
	config.AppMode = AppModeType.HandleParam
	if enable {
		config.SetSSHConfigEnabled = featureName == FeatureSSHConfig
	} else {
		config.SetSSHConfigEnabled = featureName != FeatureSSHConfig
	}
}
