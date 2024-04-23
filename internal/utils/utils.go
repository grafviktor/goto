// Package utils contains various utility methods
package utils

import (
	"errors"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strings"
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

// CheckAppInstalled - checks if application is installed and can be found in executable path
// appName - name of the application to be looked for in $PATH.
func CheckAppInstalled(appName string) error {
	_, err := exec.LookPath(appName)

	return err
}

var argumentsRegexp = regexp.MustCompile(`\s+`)

// BuildProcess - builds exec.Cmd object from command string.
func BuildProcess(cmd string) *exec.Cmd {
	if strings.TrimSpace(cmd) == "" {
		return nil
	}

	commandWithArguments := argumentsRegexp.Split(cmd, -1)
	command := commandWithArguments[0]
	arguments := commandWithArguments[1:]

	return exec.Command(command, arguments...)
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

var twoOrMoreSpacesRegexp = regexp.MustCompile(`\s{2,}`)

func RemoveDuplicateSpaces(arguments string) string {
	return twoOrMoreSpacesRegexp.ReplaceAllLiteralString(arguments, " ")
}
