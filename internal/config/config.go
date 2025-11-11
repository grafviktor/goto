// Package config contains application configuration struct
package config

import (
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/caarlos0/env/v10"
	"github.com/samber/lo"

	"github.com/grafviktor/goto/internal/ui/theme"
	"github.com/grafviktor/goto/internal/utils"
)

const (
	appName          = "goto"
	featureSSHConfig = "ssh_config"
)

type loggerInterface interface {
	Info(format string, args ...any)
	Error(format string, args ...any)
	Debug(format string, args ...any)
	Close()
}

// Configuration structs contains user-definable parameters.
type Configuration struct {
	AppName                          string
	AppHome                          string `env:"GG_HOME"`
	LogLevel                         string `env:"GG_LOG_LEVEL"            envDefault:"info"`
	SSHConfigFilePath                string `env:"GG_SSH_CONFIG_FILE_PATH"`
	IsSSHConfigFilePathDefinedByUser bool
	DisplayVersionAndExit            bool
	EnableFeature                    FeatureFlag
	DisableFeature                   FeatureFlag
	SetTheme                         string
	SetSSHConfigEnabled              bool
	AppMode                          AppMode
}

func Initialize() (*Configuration, string, error) {
	envConfig, err := parseEnvironmentVariables()
	if err != nil {
		return envConfig, "", fmt.Errorf("failed to parse environment variables: %w", err)
	}

	cmdConfig, status, err := parseCommandLineFlags(envConfig)
	if err != nil {
		return envConfig, "", err
	}

	appConfig, err := setConfigDefaults(cmdConfig)
	if err != nil {
		return envConfig, "", fmt.Errorf("failed to set configuration defaults: %w", err)
	}

	return appConfig, status, nil
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
func parseCommandLineFlags(envConfig *Configuration) (*Configuration, string, error) {
	cmdConfig := &Configuration{}

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
	var status string
	switch {
	case cmdConfig.DisplayVersionAndExit:
		status = handleDisplayVersion(cmdConfig)
	case cmdConfig.EnableFeature != "":
		status = handleFeatureToggle(cmdConfig, cmdConfig.EnableFeature.String(), true)
	case cmdConfig.DisableFeature != "":
		status = handleFeatureToggle(cmdConfig, cmdConfig.DisableFeature.String(), false)
	case cmdConfig.SetTheme != "":
		status, err = handleSetTheme(cmdConfig, cmdConfig.SetTheme)
	}

	return cmdConfig, status, err
}

func setConfigDefaults(config *Configuration) (*Configuration, error) {
	var err error

	// Set ssh config file path
	config.AppName = appName
	config.IsSSHConfigFilePathDefinedByUser = !utils.StringEmpty(&config.SSHConfigFilePath)
	config.SSHConfigFilePath, err = utils.SSHConfigFilePath(config.SSHConfigFilePath)
	if err != nil {
		return config, err
	}

	return config, nil
}

func handleDisplayVersion(config *Configuration) string {
	config.AppMode = AppModeType.DisplayInfo
	return "Display version and exit"
}

// handleFeatureToggle handles enabling or disabling features.
func handleFeatureToggle(config *Configuration, featureName string, enable bool) string {
	config.AppMode = AppModeType.HandleParam
	if enable {
		config.SetSSHConfigEnabled = featureName == featureSSHConfig
	} else {
		config.SetSSHConfigEnabled = featureName != featureSSHConfig
	}

	action := "Disable"
	if enable {
		action = "Enable"
	}

	status := fmt.Sprintf("%s feature %q and exit", action, featureName)
	log.Println(status)

	return status
}

// handleSetTheme handles setting the application theme.
func handleSetTheme(config *Configuration, themeName string) (string, error) {
	config.AppMode = AppModeType.HandleParam
	// List available themes
	availableThemes, err := theme.ListAvailableThemes(config.AppHome)
	if err != nil {
		return "", err
	}

	// Validate theme name
	themeExists := lo.Contains(availableThemes, themeName)
	if !themeExists {
		err = fmt.Errorf("theme %q not found, available themes: %q", themeName, availableThemes)
	}

	config.SetTheme = themeName
	status := fmt.Sprintf("Set theme %q and exit", themeName)
	log.Println(status)

	return status, err
}

// func (config *Configuration) IsFeatureEnabled(feature string) bool {
// 	switch feature {
// 	case featureSSHConfig:
// 		return config.SSHConfigFilePath != ""
// 	default:
// 		return false
// 	}
// }

func (config *Configuration) LogDetails(logger loggerInterface) {
	logger.Info("[CONFIG] Set application home folder to %q\n", config.AppHome)
	logger.Info("[CONFIG] Set application log level to %q\n", config.LogLevel)
	logger.Info("[CONFIG] Enabled features: %q\n", config.EnableFeature)
	logger.Info("[CONFIG] Disabled features: %q\n", config.DisableFeature)
	logger.Info("[CONFIG] Set SSH config path to %q\n", config.SSHConfigFilePath)
}
