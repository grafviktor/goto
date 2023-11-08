package logger

import (
	"fmt"
	"log"
	"os"
	"path"

	"github.com/grafviktor/goto/internal/utils"
)

type logLevel int

const (
	LevelDebug logLevel = iota
	LevelInfo
	LevelError
	LevelNone = 9
)

func New(appName string, level logLevel) (Logger, error) {
	l := Logger{}
	l.logLevel = level

	var appPath string
	appPath, err := utils.GetAppDir(&l, appName)
	if err != nil {
		return l, nil
	}

	logFilePath := path.Join(appPath, "goto.log")
	l.logFile, err = os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return l, err
	}

	l.logger = log.New(l.logFile, "", log.Ldate|log.Ltime)

	return l, nil
}

type Logger struct {
	logFile  *os.File
	logLevel logLevel
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
	if l.logLevel <= LevelError {
		l.print("ERRO", format, args...)
	}
}

func (l *Logger) Close() {
	l.logFile.Close()
}
