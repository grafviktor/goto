package logger

import (
	"context"
	"errors"
	"log"
	"os"
	"sync"

	tea "github.com/charmbracelet/bubbletea"
)

type Logger struct {
	once    sync.Once
	logFile *os.File
}

func (l *Logger) Log(format string, args ...any) error {
	var err error

	l.once.Do(func() {
		if len(os.Getenv("DEBUG")) > 0 || true { // TODO: remove force debug flag
			l.logFile, err = tea.LogToFile("debug.log", "debug")
		}
	})

	if err != nil {
		return err
	}

	log.Printf(format, args...)

	return nil
}

func (l *Logger) Close() {
	l.logFile.Close()
}

type ctxKey struct{}

// type ILogger interface {
// 	Log(v ...any) error
// 	Close()
// }

func ToContext(ctx context.Context, logger *Logger) context.Context {
	return context.WithValue(ctx, ctxKey{}, logger)
}
func FromContext(ctx context.Context) (*Logger, error) {
	if logger, ok := ctx.Value(ctxKey{}).(*Logger); ok {
		return logger, nil
	}

	return nil, errors.New("Logger not found in the context")
}
