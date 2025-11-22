// Package utils contains various utility methods
package utils //nolint:revive,nolintlint // utils is a common name

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/samber/lo"

	"github.com/grafviktor/goto/internal/constant"
)

// StringEmpty - checks if string is empty or contains only spaces.
// s is string to check.
func StringEmpty(s *string) bool {
	return s == nil || len(strings.TrimSpace(*s)) == 0
}

// FprintfIgnoreErrorf - writes formatted string to the writer, ignoring any errors.
func FprintfIgnoreErrorf(w io.Writer, format string, args ...any) {
	_, _ = fmt.Fprintf(w, format, args...)
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

// AppDir - returns application home folder where all files are stored.
// appName is application name which will be used as folder name.
// userDefinedPath allows you to set a custom path to application home folder, can be relative or absolute.
// If userDefinedPath is not empty, it will be used as application home folder
// Else, userConfigDir will be used, which is system dependent.
func AppDir(appName, userDefinedPath string) (string, error) {
	if !StringEmpty(&userDefinedPath) {
		absolutePath, err := filepath.Abs(userDefinedPath)
		if err != nil {
			return "", err
		}

		stat, err := os.Stat(absolutePath)
		if err != nil {
			return absolutePath, err
		}

		if !stat.IsDir() {
			return "", fmt.Errorf("%q is not a directory", absolutePath)
		}

		return absolutePath, nil
	}

	if StringEmpty(&appName) {
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

// IsSupportedURL checks if the given path is a URL starting with http, https, or ftp.
func IsSupportedURL(path string) bool {
	if StringEmpty(&path) {
		return false
	}

	path = strings.TrimSpace(path)
	return lo.ContainsBy([]string{"http://", "https://", "ftp://"}, func(prefix string) bool {
		return strings.HasPrefix(strings.ToLower(path), prefix)
	})
}

// ExtractBaseURL extracts the base URL (scheme + host + port) from a URL by removing the path and query parameters.
// Example: "http://127.0.0.1:8080/path/to/resource" -> "http://127.0.0.1:8080"
func ExtractBaseURL(urlPath string) (string, error) {
	if !IsSupportedURL(urlPath) {
		return "", fmt.Errorf("not supported URL format: %s", urlPath)
	}

	parsedURL, err := url.Parse(urlPath)
	if err != nil {
		return "", fmt.Errorf("invalid URL format: %w", err)
	}

	// Construct base URL with scheme, host, and port (if specified)
	baseURL := fmt.Sprintf("%s://%s", parsedURL.Scheme, parsedURL.Host)

	return baseURL, nil
}

// networkResponseTimeout - max-wait time for network requests.
// Not 'const' as redefined in unit tests to reduce execution time.
var networkResponseTimeout = 10 * time.Second

// FetchFromURL fetches content from a URL and returns it as a string.
func FetchFromURL(urlPath string) (io.ReadCloser, error) {
	if !IsSupportedURL(urlPath) {
		return nil, fmt.Errorf("not a valid URL: %s", urlPath)
	}

	parsedURL, err := url.Parse(urlPath)
	if err != nil {
		return nil, fmt.Errorf("invalid URL format: %w", err)
	}

	//nolint:noctx // want to use http.NewRequest instead of http.NewRequestWithContext
	req, err := http.NewRequest(http.MethodGet, parsedURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	client := &http.Client{Timeout: networkResponseTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		_ = resp.Body.Close()
		return nil, fmt.Errorf("failed to fetch %s: status code %d", urlPath, resp.StatusCode)
	}

	return resp.Body, nil
}

// SSHConfigFilePath - returns ssh_config path or error.
func SSHConfigFilePath(userDefinedPath string) (string, error) {
	if !StringEmpty(&userDefinedPath) {
		if IsSupportedURL(userDefinedPath) {
			return userDefinedPath, nil
		}

		// Do not expand "~".
		if strings.HasPrefix(userDefinedPath, "~") {
			return userDefinedPath, nil
		}

		absolutePath, err := filepath.Abs(userDefinedPath)
		if err != nil {
			return "", err
		}

		return absolutePath, nil
	}

	return SSHConfigDefaultFilePath()
}

// SSHConfigDefaultFilePath - returns default ssh_config path or error.
func SSHConfigDefaultFilePath() (string, error) {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s/.ssh/config", userHomeDir), nil
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
func (writer *ProcessBufferWriter) Write(p []byte) (int, error) {
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

func IsFolderExists(folderPath string) bool {
	if _, err := os.Stat(folderPath); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

var requiredBinaryInPath = "ssh"

func CheckAppRequirements(appHome string) error {
	var err error

	// Check if "ssh" utility is in application path
	if err = checkAppInstalled(requiredBinaryInPath); err != nil {
		return fmt.Errorf("%s utility is not installed or cannot be found in the executable path: %w",
			requiredBinaryInPath, err)
	}

	// Create application home folder path
	err = createAppDirIfNotExists(appHome)
	if err != nil {
		return fmt.Errorf("cannot create application home folder: %w", err)
	}

	return nil
}

// checkAppInstalled - checks if application is installed and can be found in executable path
// appName - name of the application to be looked for in $PATH.
func checkAppInstalled(appName string) error {
	_, err := exec.LookPath(appName)

	return err
}

// createAppDirIfNotExists - creates application home folder if it doesn't exist.
// appConfigDir is application home folder path.
func createAppDirIfNotExists(appConfigDir string) error {
	if StringEmpty(&appConfigDir) {
		return errors.New("bad folder name")
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

const (
	logMsgCloseApp      = "--------= Close application =-------"
	logMsgCloseAppError = "--------= Close application with non-zero code =--------"
)

type loggerInterface interface {
	Info(format string, args ...any)
	Error(format string, args ...any)
	Close()
}

// For unit tests, allows to override os.Exit.
var exitFunc = os.Exit

// LogAndCloseApp logs the close message, closes the logger, and exits with the specified code.
func LogAndCloseApp(lg loggerInterface, exitCode int, exitReason string) {
	loggingFunc := lo.Ternary(exitCode == constant.AppExitCodeSuccess, lg.Info, lg.Error)
	closeMsg := lo.Ternary(exitCode == constant.AppExitCodeSuccess, logMsgCloseApp, logMsgCloseAppError)

	if !StringEmpty(&exitReason) {
		fmt.Println(exitReason) //nolint:forbidigo // Allow printing exit reason to console
		loggingFunc(exitReason)
	}

	loggingFunc("[MAIN] %s", closeMsg)

	lg.Close()
	exitFunc(exitCode)
}
