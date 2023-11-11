package utils

import (
	"errors"
	"os"
	"os/user"
	"path"
	"path/filepath"
)

type Logger interface {
	Debug(format string, args ...any)
}

func CreateAppDirIfNotExists(appConfigDir string) error {
	_, err := os.Stat(appConfigDir)

	if os.IsNotExist(err) {
		err := os.MkdirAll(appConfigDir, 0o700)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	return nil
}

func GetAppDir(appName, userDefinedPath string) (string, error) {
	if len(userDefinedPath) > 0 {
		absolutePath, err := filepath.Abs(userDefinedPath)
		if err != nil {
			return "", err
		}

		stat, err := os.Stat(absolutePath)
		if err != nil {
			return "", err
		}

		if !stat.IsDir() {
			return "", errors.New("home path is not a directory")
		}

		return absolutePath, nil
	}

	// FIXME: -------- DEBUG ------------- /
	// userConfigDir, err := os.Getwd()
	// -------- RELEASE ----------- /
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	return path.Join(userConfigDir, appName), nil
}

func GetCurrentOSUser() string {
	user, err := user.Current()
	if err != nil {
		return "n/a"
	}

	return user.Username
}
