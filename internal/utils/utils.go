package utils

import (
	"errors"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"strings"

	"github.com/grafviktor/goto/internal/constant"
)

type Logger interface {
	Debug(format string, args ...any)
}

func stringEmpty(s string) bool {
	return len(strings.TrimSpace(s)) == 0
}

func CreateAppDirIfNotExists(appConfigDir string) error {
	if stringEmpty(appConfigDir) {
		return constant.ErrBadArgument
	}

	stat, err := os.Stat(appConfigDir)
	if os.IsNotExist(err) {
		return os.MkdirAll(appConfigDir, 0o700)
	} else if err != nil {
		return err
	}

	if !stat.IsDir() {
		return errors.New("app home path exists and it is not a directory")
	}

	return nil
}

// AppDir - returns application home folder where all all files are stored.
// appName is application name which will be used as folder name.
// userDefinedPath allows you to set a custom path to application home folder, can be relative or absolute.
// If userDefinedPath is not empty, it will be used as application home folder
// Else, userConfigDir will be used, which is system dependent.
func AppDir(appName, userDefinedPath string) (string, error) {
	if !stringEmpty(userDefinedPath) {
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

	if stringEmpty(appName) {
		return "", errors.New("application home folder name is not provided")
	}

	// Left for debugging purposes
	// userConfigDir, err := os.Getwd()
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	return path.Join(userConfigDir, appName), nil
}

// CurrentOSUsername - returns current OS username or "n/a" if it can't be determined.
func CurrentOSUsername() string {
	user, err := user.Current()
	if err != nil {
		return "n/a"
	}

	return user.Username
}
