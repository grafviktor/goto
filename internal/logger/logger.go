// Package logger incapsulates logger functions
package logger

import (
	"fmt"
	"log"
	"os"
	"path"

	"github.com/grafviktor/goto/internal/utils"
)

// LogLevel is a subject for revising. Probably it's better to have a boolean flag to switch on/off debug logging.
type LogLevel int

const (
	// LevelDebug - when log level is 'debug', then all messages will be printed into the log file.
	LevelDebug LogLevel = iota
	// LevelInfo - when log level is 'info', then only 'info' messages will be printed into the log file.
	LevelInfo
)

const logFileName = "app.log"

// New - creates a new logger with a specific log level.
// appPath - where log file will be stored.
// userSetLogLevel - user-defined log level (debug or info).
func New(appPath, userSetLogLevel string) (appLogger, error) {
	var logLevel LogLevel
	switch userSetLogLevel {
	case "debug":
		logLevel = LevelDebug
	default:
		logLevel = LevelInfo
	}

	l := appLogger{logLevel: logLevel}

	logFilePath := path.Join(appPath, logFileName)
	logFile, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return l, err
	}

	l.innerLogger = log.New(logFile, "", log.Ldate|log.Ltime)

	return l, nil
}

type appLogger struct {
	logFile     *os.File
	logLevel    LogLevel
	innerLogger *log.Logger
}

func (l *appLogger) print(prefix, format string, args ...any) {
	msg := fmt.Sprintf("[%s] %s", prefix, format)
	msg = fmt.Sprintf(msg, args...)
	l.innerLogger.Print(utils.StripStyles(msg))
}

func (l *appLogger) Debug(format string, args ...any) {
	if l.logLevel <= LevelDebug {
		l.print("DEBG", format, args...)
	}
}

func (l *appLogger) Info(format string, args ...any) {
	if l.logLevel <= LevelInfo {
		l.print("INFO", format, args...)
	}
}

func (l *appLogger) Error(format string, args ...any) {
	l.print("ERRO", format, args...)
}

func (l *appLogger) Close() {
	//nolint:errcheck // we don't care if log file wasn't closed properly
	l.logFile.Close()
}
