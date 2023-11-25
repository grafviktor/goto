package utils

import (
	"errors"
	"os"
	"os/exec"
	"os/user"
	"path"
	"path/filepath"
	"strings"

	"github.com/grafviktor/goto/internal/constant"
	"github.com/grafviktor/goto/internal/model"
	"github.com/grafviktor/goto/internal/utils/ssh"
)

// stringEmpty - checks if string is empty or contains only spaces.
// s is string to check.
func stringEmpty(s string) bool {
	return len(strings.TrimSpace(s)) == 0
}

// CreateAppDirIfNotExists - creates application home folder if it doesn't exist.
// appConfigDir is application home folder path.
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

// CurrentUsername - returns current OS username or "n/a" if it can't be determined.
func CurrentUsername() string {
	user, err := user.Current()
	if err != nil {
		return "n/a"
	}

	return user.Username
}

// CheckAppInstalled - checks if application is installed and can be found in executable path
// appName - name of the application to be looked for in $PATH
func CheckAppInstalled(appName string) error {
	_, err := exec.LookPath(appName)

	return err
}

// HostModelToOptionsAdaptor - extract values from model.Host into a set of ssh.CommandLineOption
// host - model.Host to be adapted
// returns []ssh.CommandLineOption
func HostModelToOptionsAdaptor(host model.Host) []ssh.CommandLineOption {
	return []ssh.CommandLineOption{
		ssh.OptionAddress{Value: host.Address},
		ssh.OptionLoginName{Value: host.LoginName},
		ssh.OptionRemotePort{Value: host.RemotePort},
		ssh.OptionPrivateKey{Value: host.PrivateKeyPath},
	}
}

func BuildProcess(cmd string) *exec.Cmd {
	if strings.TrimSpace(cmd) == "" {
		return nil
	}

	commandWithArguments := strings.Split(cmd, " ")
	command := commandWithArguments[0]
	arguments := commandWithArguments[1:]

	return exec.Command(command, arguments...)
}
