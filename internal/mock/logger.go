package mock

type MockLogger struct{}

func (l *MockLogger) Debug(format string, args ...any) {
}

func (l *MockLogger) Info(format string, args ...any) {
}

func (l *MockLogger) Error(format string, args ...any) {
}

func (l *MockLogger) Close() {
}
