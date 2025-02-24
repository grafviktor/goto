// Package utils contains various utility methods
package utils

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"
)

// StringEmpty - checks if string is empty or contains only spaces.
// s is string to check.
func StringEmpty(s string) bool {
	return len(strings.TrimSpace(s)) == 0
}

// Regex pattern to split at the boundary between letters and numbers.
var abbreviationRe = regexp.MustCompile(`(\p{L}+|\p{N}+|\p{S}+)`)

// StringAbbreviation - creates an abbreviation from a string, combining starting letters from first and last words.
// For example:
//
//	"Alexandria, Egypt"      -> "AE"
//	"Babylon Iraq"           -> "BI"
//	"Carthage, North Africa" -> "CA"
//	"Thebes_Greece"          -> "TG"
func StringAbbreviation(s string) string {
	parts := abbreviationRe.FindAllString(s, -1)
	// If there is more than one word, create abbreviation from first and last words
	if len(parts) > 1 {
		wordFirst := []rune(parts[0])
		wordLast := []rune(parts[len(parts)-1])

		return fmt.Sprintf("%c%c", unicode.ToUpper(wordFirst[0]), unicode.ToUpper(wordLast[0]))
	}

	// If there is single word only, attempt to build abbreviation assuming it's
	// in camelCase. Otherwise just fallback on the first letter of the word.
	if len(parts) == 1 {
		word := []rune(parts[0])
		var letterFirst, letterSecond rune
		for i, r := range word {
			if i == 0 {
				letterFirst = unicode.ToUpper(r)
				letterSecond = ' '
				continue
			}

			if unicode.IsUpper(r) {
				letterSecond = r
			}
		}

		result := fmt.Sprintf("%c%c", letterFirst, letterSecond)
		return strings.TrimSpace(result)
	}

	return ""
}

var ansiRegex = regexp.MustCompile("\x1b\\[[0-9;]*m")

// StripStyles - removes lipgloss styles from a string.
func StripStyles(input string) string {
	input = strings.TrimSpace(input)
	return ansiRegex.ReplaceAllString(input, "")
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

// splitArguments - converts a command with arguments into an array of strings.
// Note, that it does not preserves inner quote characters:
//
//	ssh -o option="123 456"
//	// will be split into 3 this array:
//	"ssh" "-o" "option=123 456" // no quotes around 123 456
func splitArguments(cmd string) []string {
	args := make([]string, 0)
	inQuotes := false
	commandLength := len(cmd)

	var arg string
	for charIndex, ch := range cmd {
		isQuoteCharacter := ch == '"' || ch == '\''
		isSpaceCharacter := ch == ' '

		switch {
		case isSpaceCharacter && !inQuotes:
			args = append(args, arg)
			arg = ""
		case isQuoteCharacter:
			inQuotes = !inQuotes
		default:
			arg += string(ch)
		}

		isLastCharacter := charIndex == commandLength-1
		if isLastCharacter {
			args = append(args, arg)
		}
	}

	return args
}

// BuildProcess - builds exec.Cmd object from command string.
func BuildProcess(cmd string) *exec.Cmd {
	if strings.TrimSpace(cmd) == "" {
		return nil
	}

	commandWithArguments := splitArguments(cmd)
	command := commandWithArguments[0]
	arguments := commandWithArguments[1:]

	return exec.Command(command, arguments...)
}

// ProcessBufferWriter - is an object which pretends to be a writer, however it saves all data into a temporary buffer
// variable for future reading and doesn't write anything in terminal. Utilized to parse process stdout or stderr.
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

// RemoveDuplicateSpaces - removes two or more spaces from the string.
func RemoveDuplicateSpaces(arguments string) string {
	return twoOrMoreSpacesRegexp.ReplaceAllLiteralString(arguments, " ")
}

// BuildProcessInterceptStdErr - builds a process where stderr is intercepted for further processing.
func BuildProcessInterceptStdErr(command string) *exec.Cmd {
	process := BuildProcess(command)
	process.Stdout = os.Stdout
	process.Stderr = &ProcessBufferWriter{}

	return process
}

// BuildProcessInterceptStdAll - builds a process where both stdout and stderr are intercepted for further processing.
func BuildProcessInterceptStdAll(command string) *exec.Cmd {
	// Use case 1: User edits host
	// Use case 2: User is going to copy his ssh key using <t> command from the hostlist

	process := BuildProcess(command)
	process.Stdout = &ProcessBufferWriter{}
	process.Stderr = &ProcessBufferWriter{}

	return process
}
