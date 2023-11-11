package logger

import (
	"fmt"
	"log"
	"os"
	"path"
)

// LogLevel is a subject for revising. Probably it's better to have a boolean flag to switch on/off debug logging
type LogLevel int

const (
	LevelDebug LogLevel = iota
	LevelInfo
)

const logFileName = "app.log"

func New(appPath, userSetLogLevel string) (Logger, error) {
	var logLevel LogLevel
	switch userSetLogLevel {
	case "debug":
		logLevel = LevelDebug
	default:
		logLevel = LevelInfo
	}

	l := Logger{logLevel: logLevel}

	logFilePath := path.Join(appPath, logFileName)
	logFile, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return l, err
	}

	l.logger = log.New(logFile, "", log.Ldate|log.Ltime)

	return l, nil
}

type Logger struct {
	logFile  *os.File
	logLevel LogLevel
	logger   *log.Logger
}

func (l *Logger) print(prefix, format string, args ...any) {
	msg := fmt.Sprintf("[%s] %s", prefix, format)
	l.logger.Printf(msg, args...)
}

func (l *Logger) Debug(format string, args ...any) {
	if l.logLevel <= LevelDebug {
		l.print("DEBG", format, args...)
	}
}

func (l *Logger) Info(format string, args ...any) {
	if l.logLevel <= LevelInfo {
		l.print("INFO", format, args...)
	}
}

func (l *Logger) Error(format string, args ...any) {
	l.print("ERRO", format, args...)
}

func (l *Logger) Close() {
	l.logFile.Close()
}
