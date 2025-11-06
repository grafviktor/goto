package application

import (
	"flag"
	"fmt"
	"strings"

	"github.com/caarlos0/env/v10"
	"github.com/grafviktor/goto/internal/state"
	"github.com/grafviktor/goto/internal/ui/theme"
	"github.com/grafviktor/goto/internal/utils"
	"github.com/samber/lo"
)

const (
	appName          = "goto"
	defaultThemeName = "default"
	featureSSHConfig = "ssh_config"
)

func Parse() (*Configuration, error) {
	envConfig, err := parseEnvironmentConfig()
	if err != nil {
		// fmt.Printf("Error parsing environment configuration: %+v\n", err)
		return nil, fmt.Errorf("error parsing environment configuration: %w", err)
	}

	// Check if "ssh" utility is in application path
	if err = utils.CheckAppInstalled("ssh"); err != nil {
		// log.Fatalf("[MAIN] ssh utility is not installed or cannot be found in the executable path: %v", err)
		return nil, fmt.Errorf("ssh utility is not installed or cannot be found in the executable path: %w", err)
	}

	cmdConfig := parseCommandLineFlags(envConfig)
	return setupApplicationConfiguration(cmdConfig)
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

	return cmdConfig
}

type loggerInterface interface {
	Info(format string, args ...any)
	Error(format string, args ...any)
	Debug(format string, args ...any)
	Close()
}

// HandleFeatureToggle handles enabling or disabling features.
func HandleFeatureToggle(lg loggerInterface, appState *state.Application, featureName string, enable bool) error {
	action := "Disable"
	if enable {
		action = "Enable"
	}

	lg.Info("[MAIN] %s feature %q and exit", action, featureName)
	fmt.Printf("%sd: '%s'\n", action, featureName)

	if enable {
		appState.SSHConfigEnabled = featureName == featureSSHConfig
	} else {
		appState.SSHConfigEnabled = featureName != featureSSHConfig
	}

	return appState.Persist()
}

// HandleSetTheme handles setting the application theme.
func HandleSetTheme(lg loggerInterface, appState *state.Application, themeName string) error {
	// List available themes
	availableThemes, err := theme.ListAvailableThemes(appState.ApplicationConfig.AppHome, lg)
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
	appState.Theme = themeName
	lg.Info("[CONFIG] Set theme to %q", themeName)
	fmt.Printf("Theme set to: '%s'\n", themeName)

	return appState.Persist()
}

// setupApplicationConfiguration sets up the application configuration and validates it.
func setupApplicationConfiguration(config Configuration) (*Configuration, error) {
	// success := true
	var err error

	// Set application home folder path
	config.AppHome, err = utils.AppDir(appName, config.AppHome)
	if err != nil {
		// log.Printf("[MAIN] Application home folder error: %v", err)
		// success = false
		// return nil, fmt.Errorf("application home folder error: %w", err)
		err = utils.CreateAppDirIfNotExists(config.AppHome)
		if err != nil {
			return nil, fmt.Errorf("cannot create or access application home folder error: %w", err)
		}
	}

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
		return nil, fmt.Errorf("cannot open SSH config file: %w", err)
	}

	return &config, nil
}
