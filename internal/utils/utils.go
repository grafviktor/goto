package utils

import (
	"errors"
	"os"
	"os/exec"
	"os/user"
	"path"
	"path/filepath"
	"strings"

	"github.com/grafviktor/goto/internal/model"
	"github.com/grafviktor/goto/internal/utils/ssh"
)

type Logger interface {
	Debug(format string, args ...any)
}

func CreateAppDirIfNotExists(appConfigDir string) error {
	_, err := os.Stat(appConfigDir)

	if os.IsNotExist(err) {
		err = os.MkdirAll(appConfigDir, 0o700)
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
