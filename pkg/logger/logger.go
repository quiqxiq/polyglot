package logger

import (
	"context"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
)

var (
	defaultLogger *logrus.Logger
	once          sync.Once
)

func init() {
	once.Do(func() {
		defaultLogger = logrus.New()
		defaultLogger.SetOutput(os.Stdout)
		defaultLogger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05.000",
			ForceColors:     true,
		})
		defaultLogger.SetLevel(logrus.InfoLevel)
	})
}

// Init configures the global logger based on environment and log level.
func Init(level, env string, output ...io.Writer) {
	if len(output) > 0 && output[0] != nil {
		defaultLogger.SetOutput(output[0])
	} else {
		defaultLogger.SetOutput(os.Stdout)
	}

	// Configure formatter based on environment
	if strings.ToLower(env) == "production" || strings.ToLower(env) == "prod" {
		defaultLogger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02T15:04:05.000Z07:00",
		})
	} else {
		defaultLogger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05.000",
			ForceColors:     true,
		})
	}

	// Parse and set log level
	parsedLevel, err := logrus.ParseLevel(level)
	if err != nil {
		defaultLogger.SetLevel(logrus.InfoLevel)
	} else {
		defaultLogger.SetLevel(parsedLevel)
	}
}

// GetLogger returns the underlying logrus.Logger instance.
func GetLogger() *logrus.Logger {
	return defaultLogger
}

// WithComponent creates an Entry tagged with a component name.
func WithComponent(component string) *logrus.Entry {
	return defaultLogger.WithField("component", component)
}

// WithField adds a single field to the log entry.
func WithField(key string, value any) *logrus.Entry {
	return defaultLogger.WithField(key, value)
}

// WithFields adds multiple fields to the log entry.
func WithFields(fields logrus.Fields) *logrus.Entry {
	return defaultLogger.WithFields(fields)
}

// WithContext returns a log entry with context attached.
func WithContext(ctx context.Context) *logrus.Entry {
	return defaultLogger.WithContext(ctx)
}

// WithError returns a log entry with an error attached.
func WithError(err error) *logrus.Entry {
	return defaultLogger.WithError(err)
}

// Info logs a message at level Info on the standard logger.
func Info(args ...any) {
	defaultLogger.Info(args...)
}

// Infof logs a message at level Info on the standard logger.
func Infof(format string, args ...any) {
	defaultLogger.Infof(format, args...)
}

// Debug logs a message at level Debug on the standard logger.
func Debug(args ...any) {
	defaultLogger.Debug(args...)
}

// Debugf logs a message at level Debug on the standard logger.
func Debugf(format string, args ...any) {
	defaultLogger.Debugf(format, args...)
}

// Warn logs a message at level Warn on the standard logger.
func Warn(args ...any) {
	defaultLogger.Warn(args...)
}

// Warnf logs a message at level Warn on the standard logger.
func Warnf(format string, args ...any) {
	defaultLogger.Warnf(format, args...)
}

// Error logs a message at level Error on the standard logger.
func Error(args ...any) {
	defaultLogger.Error(args...)
}

// Errorf logs a message at level Error on the standard logger.
func Errorf(format string, args ...any) {
	defaultLogger.Errorf(format, args...)
}

// Fatal logs a message at level Fatal on the standard logger then exits.
func Fatal(args ...any) {
	defaultLogger.Fatal(args...)
}

// Fatalf logs a message at level Fatal on the standard logger then exits.
func Fatalf(format string, args ...any) {
	defaultLogger.Fatalf(format, args...)
}
