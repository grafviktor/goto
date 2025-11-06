package application

import (
	"flag"
	"fmt"
	"strings"

	"github.com/caarlos0/env/v10"
	"github.com/grafviktor/goto/internal/ui/theme"
	"github.com/grafviktor/goto/internal/utils"
	"github.com/samber/lo"
)

const (
	appName          = "goto"
	defaultThemeName = "default"
	featureSSHConfig = "ssh_config"
)

type loggerInterface interface {
	Info(format string, args ...any)
	Error(format string, args ...any)
	Debug(format string, args ...any)
	Close()
}

func Create() (*Configuration, error) {
	envConfig, err := parseEnvironmentVariables()
	if err != nil {
		return nil, fmt.Errorf("failed to parse environment variables: %w", err)
	}

	cmdConfig := parseCommandLineFlags(envConfig)

	appConfig, err := setConfigDefaults(cmdConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to set configuration defaults: %w", err)
	}

	return &appConfig, nil
}

func parseEnvironmentVariables() (Configuration, error) {
	envConfig, err := parseEnvironmentConfig()
	if err != nil {
		// fmt.Printf("Error parsing environment configuration: %+v\n", err)
		return envConfig, fmt.Errorf("error parsing environment configuration: %w", err)
	}

	return envConfig, nil
}

// parseEnvironmentConfig parses environment configuration.
func parseEnvironmentConfig() (Configuration, error) {
	envConfig := Configuration{}
	err := env.Parse(&envConfig)
	return envConfig, err
}

// parseCommandLineFlags parses command line flags and returns the configuration.
func parseCommandLineFlags(envConfig Configuration) Configuration {
	cmdConfig := Configuration{}

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

	switch {
	case cmdConfig.EnableFeature != "":
		cmdConfig.ShouldExitAfterConfig = true
	case cmdConfig.DisableFeature != "":
		cmdConfig.ShouldExitAfterConfig = true
	case cmdConfig.SetTheme != "":
		cmdConfig.ShouldExitAfterConfig = true
	default:
		cmdConfig.ShouldExitAfterConfig = false
	}

	return cmdConfig
}

func setConfigDefaults(config Configuration) (Configuration, error) {
	// success := true
	var err error

	// Create application folder
	// if err = utils.CreateAppDirIfNotExists(config.AppHome); err != nil {
	// 	log.Printf("[MAIN] Can't create application home folder: %v", err)
	// } else {
	// 	// Even if there was an error, we created the application home folder.
	// 	success = true
	// }

	// Set ssh config file path
	config.IsSSHConfigFilePathDefinedByUser = !utils.StringEmpty(&config.SSHConfigFilePath)
	config.SSHConfigFilePath, err = utils.SSHConfigFilePath(config.SSHConfigFilePath)
	if err != nil {
		// log.Printf("[MAIN] Can't open SSH config file: %v", err)
		// success = false
		return config, fmt.Errorf("cannot open SSH config file: %w", err)
	}

	return config, nil
}

// HandleFeatureToggle handles enabling or disabling features.
func HandleFeatureToggle(lg loggerInterface, config *Configuration, featureName string, enable bool) {
	if enable {
		config.SetSSHConfigEnabled = featureName == featureSSHConfig
	} else {
		config.SetSSHConfigEnabled = featureName != featureSSHConfig
	}

	action := "Disable"
	if enable {
		action = "Enable"
	}

	lg.Info("[MAIN] %s feature %q and exit", action, featureName)
	fmt.Printf("%sd: '%s'\n", action, featureName)
}

// HandleSetTheme handles setting the application theme.
func HandleSetTheme(lg loggerInterface, config *Configuration, themeName string) error {
	// List available themes
	availableThemes, err := theme.ListAvailableThemes(config.AppHome, lg)
	if err != nil {
		lg.Error("[MAIN] Cannot list available themes: %v.", err)
		availableThemes = []string{defaultThemeName}
	}

	// Validate theme name
	themeExists := lo.Contains(availableThemes, themeName)
	if !themeExists {
		lg.Error("[MAIN] Theme %q not found", themeName)
		// logCloseAndExit(lg, exitCodeError, logMessage)
		return fmt.Errorf("theme %q not found. Available themes: %q", themeName, availableThemes)
	}

	// Set theme in application state
	config.SetTheme = themeName
	lg.Info("[CONFIG] Set theme to %q", themeName)
	fmt.Printf("Theme set to: '%s'\n", themeName)

	return nil
}
