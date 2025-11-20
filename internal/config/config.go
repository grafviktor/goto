// Package config contains application configuration struct
//
//nolint:forbidigo // Use fmt.Printf to display application messages.
package config

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/caarlos0/env/v10"
	"github.com/samber/lo"

	"github.com/grafviktor/goto/internal/constant"
	"github.com/grafviktor/goto/internal/utils"
)

const (
	appName          = "goto"
	FeatureSSHConfig = "ssh_config"
)

// Configuration structs contains user-definable parameters.
type Configuration struct {
	AppMode           constant.AppMode
	AppName           string
	DisableFeature    FeatureFlag
	EnableFeature     FeatureFlag
	SetTheme          string
	AppHome           string            `env:"GG_HOME"`
	LogLevel          constant.LogLevel `env:"GG_LOG_LEVEL"            envDefault:"info"`
	SSHConfigFilePath string            `env:"GG_SSH_CONFIG_FILE_PATH"`
}

func Initialize() (*Configuration, error) {
	envConfig, err := parseEnvironmentVariables()
	if err != nil {
		return envConfig, fmt.Errorf("failed to parse environment variables: %w", err)
	}

	cmdConfig, err := parseCommandLineFlags(envConfig, os.Args)
	if err != nil {
		return nil, fmt.Errorf("failed to parse command line flags: %w", err)
	}

	appConfig, err := setConfigDefaults(cmdConfig)

	return appConfig, err
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
func parseCommandLineFlags(envConfig *Configuration, args []string) (*Configuration, error) {
	var cmdConfig Configuration
	var shouldDisplayVersionAndExit bool

	fs := flag.NewFlagSet(appName, flag.ContinueOnError)
	// Command line parameters have the highest precedence, use envConfig as fallback values
	fs.BoolVar(&shouldDisplayVersionAndExit, "v", false, "Display application details")
	fs.StringVar(&cmdConfig.AppHome, "f", envConfig.AppHome, "Application home folder")
	fs.StringVar(&cmdConfig.LogLevel, "l", envConfig.LogLevel, "Log verbosity level: debug, info")
	fs.StringVar(
		&cmdConfig.SSHConfigFilePath,
		"s",
		envConfig.SSHConfigFilePath,
		"Specifies an alternative per-user SSH configuration file path",
	)
	fs.Var(
		&cmdConfig.EnableFeature,
		"e",
		fmt.Sprintf("Enable feature. Supported values: %s", strings.Join(SupportedFeatures, "|")),
	)
	fs.Var(
		&cmdConfig.DisableFeature,
		"d",
		fmt.Sprintf("Disable feature. Supported values: %s", strings.Join(SupportedFeatures, "|")),
	)
	fs.StringVar(&cmdConfig.SetTheme, "set-theme", "", "Set application theme")

	err := fs.Parse(args)
	if err != nil {
		return nil, err
	}

	switch {
	case shouldDisplayVersionAndExit:
		cmdConfig.AppMode = constant.AppModeType.DisplayInfo
	case cmdConfig.EnableFeature != "":
		fmt.Printf("[CONFIG] Enable feature %q\n", cmdConfig.EnableFeature.String())
		cmdConfig.AppMode = constant.AppModeType.HandleParam
	case cmdConfig.DisableFeature != "":
		fmt.Printf("[CONFIG] Disable feature %q\n", cmdConfig.DisableFeature.String())
		cmdConfig.AppMode = constant.AppModeType.HandleParam
	case cmdConfig.SetTheme != "":
		fmt.Printf("[CONFIG] Set theme to %q\n", cmdConfig.SetTheme)
		cmdConfig.AppMode = constant.AppModeType.HandleParam
	}

	return &cmdConfig, nil
}

func setConfigDefaults(config *Configuration) (*Configuration, error) {
	var err error
	config.AppName = appName
	config.AppHome, err = utils.AppDir(appName, config.AppHome)
	if err != nil {
		return nil, fmt.Errorf("application home folder error: %w", err)
	}

	if utils.StringEmpty(&config.AppMode) {
		config.AppMode = constant.AppModeType.StartUI
	}

	supportedLogLevels := []constant.LogLevel{constant.LogLevelType.DEBUG, constant.LogLevelType.INFO}
	if !lo.Contains(supportedLogLevels, config.LogLevel) {
		return nil, fmt.Errorf("unsupported log level: %q", config.LogLevel)
	}

	return config, nil
}
