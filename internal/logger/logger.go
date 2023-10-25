package logger

import (
	"context"
	"errors"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func New() (Logger, error) {
	var l Logger
	var err error

	if len(os.Getenv("DEBUG")) > 0 || true { // TODO: remove force debug flag
		l.logFile, err = tea.LogToFile("debug.log", "debug")
	}

	if err != nil {
		return l, err
	}

	return Logger{}, nil
}

type Logger struct {
	logFile *os.File
}

func (l *Logger) Debug(format string, args ...any) {
	log.Printf(format, args...)
}

func (l *Logger) Close() {
	l.logFile.Close()
}

type ctxKey struct{}

func ToContext(ctx context.Context, logger *Logger) context.Context {
	return context.WithValue(ctx, ctxKey{}, logger)
}

func FromContext(ctx context.Context) (*Logger, error) {
	if logger, ok := ctx.Value(ctxKey{}).(*Logger); ok {
		return logger, nil
	}

	return nil, errors.New("Logger not found in the context")
}
