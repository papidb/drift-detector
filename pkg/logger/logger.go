package logger

import (
	"github.com/sirupsen/logrus"
)

// Logger is a custom logging interface.
type Logger interface {
	Debug(args ...any)
	Info(args ...any)
	Warn(args ...any)
	Error(args ...any)
}

type Fields map[string]interface{}

// logrusLogger implements the Logger interface using logrus.
type logrusLogger struct {
	logger *logrus.Logger
}

func Init() {
	logrus.SetLevel(logrus.DebugLevel) // Set default level here
}

// NewLogger initializes and returns a Logger implementation.
func NewLogger() Logger {
	log := logrus.New()
	return &logrusLogger{logger: log}
}

func (l *logrusLogger) Debug(args ...any) {
	if len(args) > 0 {
		if fields, ok := args[0].(Fields); ok {
			l.logger.WithFields(logrus.Fields(fields)).Debug(args[1:]...)
			return
		}
	}
	l.logger.Debug(args...)
}

func (l *logrusLogger) Info(args ...any) {
	if len(args) > 0 {
		if fields, ok := args[0].(Fields); ok {
			l.logger.WithFields(logrus.Fields(fields)).Info(args[1:]...)
			return
		}
	}
	l.logger.Info(args...)
}

func (l *logrusLogger) Warn(args ...any) {
	if len(args) > 0 {
		if fields, ok := args[0].(Fields); ok {
			l.logger.WithFields(logrus.Fields(fields)).Warn(args[1:]...)
			return
		}
	}
	l.logger.Warn(args...)
}

func (l *logrusLogger) Error(args ...any) {
	if len(args) > 0 {
		if fields, ok := args[0].(Fields); ok {
			l.logger.WithFields(logrus.Fields(fields)).Error(args[1:]...)
			return
		}
	}
	l.logger.Error(args...)
}
