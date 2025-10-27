package temporal

import (
	"github.com/sirupsen/logrus"
	"go.temporal.io/sdk/log"
)

// temporalLogger is an adapter that wraps logrus.Logger to implement Temporal's log.Logger interface
type temporalLogger struct {
	logger *logrus.Logger
}

// NewTemporalLogger creates a new Temporal logger adapter from a logrus logger
func NewTemporalLogger(logger *logrus.Logger) log.Logger {
	return &temporalLogger{
		logger: logger,
	}
}

// Debug logs a message at Debug level
func (l *temporalLogger) Debug(msg string, keyvals ...interface{}) {
	l.logger.WithFields(extractFields(keyvals)).Debug(msg)
}

// Info logs a message at Info level
func (l *temporalLogger) Info(msg string, keyvals ...interface{}) {
	l.logger.WithFields(extractFields(keyvals)).Info(msg)
}

// Warn logs a message at Warn level
func (l *temporalLogger) Warn(msg string, keyvals ...interface{}) {
	l.logger.WithFields(extractFields(keyvals)).Warn(msg)
}

// Error logs a message at Error level
func (l *temporalLogger) Error(msg string, keyvals ...interface{}) {
	l.logger.WithFields(extractFields(keyvals)).Error(msg)
}

// extractFields converts key-value pairs to logrus.Fields
func extractFields(keyvals []interface{}) logrus.Fields {
	fields := logrus.Fields{}
	for i := 0; i < len(keyvals); i += 2 {
		if i+1 < len(keyvals) {
			key, ok := keyvals[i].(string)
			if !ok {
				continue
			}
			fields[key] = keyvals[i+1]
		}
	}
	return fields
}
