package mocklogger

import "fmt"

type Logger struct {
	Logs []string
}

func (l *Logger) print(format string, args ...interface{}) {
	logMessage := format
	if len(args) > 0 {
		logMessage = fmt.Sprintf(format, args...)
	}
	l.Logs = append(l.Logs, logMessage)
}

func (l *Logger) Debug(format string, args ...any) {
	l.print(format, args...)
}

func (l *Logger) Info(format string, args ...any) {
	l.print(format, args...)
}

func (l *Logger) Warn(format string, args ...any) {
	l.print(format, args...)
}

func (l *Logger) Error(format string, args ...any) {
	l.print(format, args...)
}

func (l *Logger) Close() {
}
