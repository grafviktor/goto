package testutils_test

import "fmt"

type MockLogger struct {
	Logs []string
}

func (ml *MockLogger) print(format string, args ...interface{}) {
	logMessage := format
	if len(args) > 0 {
		logMessage = fmt.Sprintf(format, args...)
	}
	ml.Logs = append(ml.Logs, logMessage)
}

func (l *MockLogger) Debug(format string, args ...any) {
	l.print(format, args...)
}

func (l *MockLogger) Info(format string, args ...any) {
	l.print(format, args...)
}

func (l *MockLogger) Error(format string, args ...any) {
	l.print(format, args...)
}

func (l *MockLogger) Close() {
}
