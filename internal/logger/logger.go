// Package logger incapsulates logger functions
package logger

import (
	"fmt"
	"log"
	"os"
	"path"
	"sync"

	"github.com/samber/lo"

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

// This interface is used to avoid sync.Once restrictions in unit-tests.
type Once interface {
	Do(func())
}

var (
	appLogger   *AppLogger
	once        Once = &sync.Once{}
	logFileName      = "app.log"
)

// Create - creates a new logger with a specific log level.
// appPath - where log file will be stored.
// userSetLogLevel - user-defined log level (debug or info).
func Create(appPath, userSetLogLevel string) (*AppLogger, error) {
	var err error
	once.Do(func() {
		logLevel := lo.Ternary(userSetLogLevel == "debug", LevelDebug, LevelInfo)
		appLogger = &AppLogger{logLevel: logLevel}
		logFilePath := path.Join(appPath, logFileName)
		logFile, openLogFileError := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
		if openLogFileError != nil {
			log.Printf("[MAIN] Can't create application logger: %v\n", openLogFileError)
			err = openLogFileError
			return
		}

		appLogger.innerLogger = log.New(logFile, "", log.Ldate|log.Ltime)
	})

	return appLogger, err
}

// Get - returns application state.
func Get() *AppLogger {
	return appLogger
}

type AppLogger struct {
	logFile     *os.File
	logLevel    LogLevel
	innerLogger *log.Logger
}

func (l *AppLogger) print(prefix, format string, args ...any) {
	msg := fmt.Sprintf("[%s] %s", prefix, format)
	msg = fmt.Sprintf(msg, args...)
	l.innerLogger.Print(utils.StripStyles(msg))
}

func (l *AppLogger) Debug(format string, args ...any) {
	if l.logLevel <= LevelDebug {
		l.print("DEBG", format, args...)
	}
}

func (l *AppLogger) Info(format string, args ...any) {
	if l.logLevel <= LevelInfo {
		l.print("INFO", format, args...)
	}
}

func (l *AppLogger) Warn(format string, args ...any) {
	l.print("WARN", format, args...)
}

func (l *AppLogger) Error(format string, args ...any) {
	l.print("ERRO", format, args...)
}

func (l *AppLogger) Close() {
	//nolint:errcheck // we don't care if log file wasn't closed properly
	l.logFile.Close()
}
