package utils

import (
	"os"
	"os/user"
	"path"
)

type Logger interface {
	Debug(format string, args ...any)
}

func createAppDir(logger Logger, appConfigDir string) error {
	err := os.MkdirAll(appConfigDir, 0o700)
	if err != nil {
		logger.Debug("Error %s", err.Error())
		return err
	}

	return nil
}

func GetAppDir(logger Logger, appName string) (string, error) {
	// FIXME: -------- DEBUG ------------- /
	// userConfigDir, err := os.Getwd()
	// -------- RELEASE ----------- /
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	appConfigDir := path.Join(userConfigDir, appName)
	_, err = os.Stat(appConfigDir)

	if os.IsNotExist(err) {
		if err = createAppDir(logger, appConfigDir); err != nil {
			return "", err
		}
	} else if err != nil {
		logger.Debug("Failed to open or create App home folder %s\n", appConfigDir)
		return "", err
	}

	return appConfigDir, nil
}

func GetCurrentOSUser() string {
	user, err := user.Current()
	if err != nil {
		return "n/a"
	}

	return user.Username
}
