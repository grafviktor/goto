package state

import (
	"context"

	"github.com/grafviktor/goto/internal/application"
)

const (
	appName          = "goto"
	defaultThemeName = "default"
	featureSSHConfig = "ssh_config"
	// logMsgCloseApp      = "--------= Close application =-------"
	// logMsgCloseAppError = "--------= Close application with non-zero code =--------"
	// exitCodeError       = 1
	// exitCodeSuccess     = 0
)

func Initialize(
	applicationConfiguration *application.Configuration,
	lg loggerInterface,
) (*Application, error) {
	// Create application configuration
	// applicationConfiguration, applicationConfigErr := createConfigurationOrExit()

	// Create application logger.
	// lg, err := logger.Create(applicationConfiguration.AppHome, applicationConfiguration.LogLevel)
	// if err != nil {
	// 	return nil, fmt.Errorf("cannot create application logger: %w", err)
	// }

	// Create application state
	applicationState := Create(context.Background(), *applicationConfiguration, lg)

	return applicationState, nil
}
