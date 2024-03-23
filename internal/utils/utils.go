// Package utils contains various utility methods
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

// StringEmpty - checks if string is empty or contains only spaces.
// s is string to check.
func StringEmpty(s string) bool {
	return len(strings.TrimSpace(s)) == 0
}

// CreateAppDirIfNotExists - creates application home folder if it doesn't exist.
// appConfigDir is application home folder path.
func CreateAppDirIfNotExists(appConfigDir string) error {
	if StringEmpty(appConfigDir) {
		return errors.New("bad argument")
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

// AppDir - returns application home folder where all files are stored.
// appName is application name which will be used as folder name.
// userDefinedPath allows you to set a custom path to application home folder, can be relative or absolute.
// If userDefinedPath is not empty, it will be used as application home folder
// Else, userConfigDir will be used, which is system dependent.
func AppDir(appName, userDefinedPath string) (string, error) {
	if !StringEmpty(userDefinedPath) {
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

	if StringEmpty(appName) {
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
	// Read from 'ssh -G hostname' output:
	// 1. 'identityfile'
	// 2. 'user'
	// 3. 'port'
	// Example:
	// ssh -G localhost
	//
	// user roman
	// hostname localhost
	// port 22
	// identityfile ~/.ssh/id_rsa
	// identityfile ~/.ssh/id_dsa
	// identityfile ~/.ssh/id_ecdsa
	// identityfile ~/.ssh/id_ecdsa_sk
	// identityfile ~/.ssh/id_ed25519
	// identityfile ~/.ssh/id_ed25519_sk
	// identityfile ~/.ssh/id_xmss

	// That's a naive implementation. ssh [-vvv] -G <hostname> should be used to request settings for a hostname.
	user, err := user.Current()
	if err != nil {
		return "n/a"
	}

	return user.Username
}

// CheckAppInstalled - checks if application is installed and can be found in executable path
// appName - name of the application to be looked for in $PATH.
func CheckAppInstalled(appName string) error {
	_, err := exec.LookPath(appName)

	return err
}

// BuildProcess - builds exec.Cmd object from command string.
func BuildProcess(cmd string) *exec.Cmd {
	if strings.TrimSpace(cmd) == "" {
		return nil
	}

	commandWithArguments := strings.Split(cmd, " ")
	command := commandWithArguments[0]
	arguments := commandWithArguments[1:]

	return exec.Command(command, arguments...)
}

// =============================== Move to SSH module:

// HostModelToOptionsAdaptor - extract values from model.Host into a set of ssh.CommandLineOption
// host - model.Host to be adapted
// returns []ssh.CommandLineOption.
func HostModelToOptionsAdaptor(host model.Host) []ssh.CommandLineOption {
	return []ssh.CommandLineOption{
		ssh.OptionPrivateKey{Value: host.PrivateKeyPath},
		ssh.OptionRemotePort{Value: host.RemotePort},
		ssh.OptionLoginName{Value: host.LoginName},
		ssh.OptionAddress{Value: host.Address},
	}
}

// BuildConnectSSH - builds ssh command which is based on host.Model.
func BuildConnectSSH(host model.Host) *exec.Cmd {
	command := ssh.ConstructCMD(ssh.BaseCMD(), HostModelToOptionsAdaptor(host)...)
	process := BuildProcess(command)
	process.Stdout = os.Stdout
	process.Stderr = &ProcessBufferWriter{}

	return process
}

// BuildLoadSSHConfig - builds ssh command, which runs ssh -G <hostname> command
// to get a list of options associated with the hostname.
func BuildLoadSSHConfig(hostname string) *exec.Cmd {
	// Usecase 1: User edits host
	// Usecase 2: User is going to copy his ssh key using <t> command from the hostlist

	command := ssh.ConstructCMD(ssh.BaseCMD(), ssh.OptionReadConfig{Value: hostname})
	process := BuildProcess(command)
	process.Stdout = &ProcessBufferWriter{}
	process.Stderr = &ProcessBufferWriter{}

	return process
}

// ProcessBufferWriter - is an object which pretends to be a writer, however it saves all data into 'Output' variable
// for future reading and do not write anything in terminal. We need it to display or parse process output or error.
type ProcessBufferWriter struct {
	Output []byte
}

// Write - doesn't write anything, it saves all data in err variable, which can ve read later.
func (writer *ProcessBufferWriter) Write(p []byte) (n int, err error) {
	writer.Output = append(writer.Output, p...)

	// Hide output from the console, otherwise it will be seen in a subsequent ssh calls
	// To return to default behavior use: return os.{Stderr|Stdout}.Write(p)
	// We must return the number of bytes which were written using `len(p)`,
	// otherwise exec.go will throw 'short write' error.
	return len(p), nil
}
