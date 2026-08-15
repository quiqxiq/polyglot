package logger

import (
	"context"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// Fields is a type alias for logrus.Fields to avoid external imports in caller files.
type Fields = logrus.Fields

// Level is a type alias for logrus.Level.
type Level = logrus.Level

// Entry is a type alias for logrus.Entry.
type Entry = logrus.Entry

// Logger is a type alias for logrus.Logger.
type Logger = logrus.Logger

var (
	globalLogger *logrus.Logger
	once         sync.Once
	mu           sync.RWMutex
)

func init() {
	globalLogger = logrus.New()
	globalLogger.SetOutput(os.Stdout)
	globalLogger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: time.RFC3339,
		ForceColors:     true,
	})
	globalLogger.SetLevel(logrus.InfoLevel)
}

// Init configures the global Logrus logger instance.
// appEnv: "production", "staging", or "development".
// levelStr: "trace", "debug", "info", "warn", "error", "fatal", "panic".
func Init(appEnv string, levelStr string, out io.Writer) *logrus.Logger {
	mu.Lock()
	defer mu.Unlock()

	if out == nil {
		out = os.Stdout
	}
	globalLogger.SetOutput(out)

	// Configure Formatter
	isProd := strings.EqualFold(appEnv, "production") || strings.EqualFold(appEnv, "staging")
	if isProd {
		globalLogger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: time.RFC3339,
		})
	} else {
		globalLogger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: time.RFC3339,
			ForceColors:     true,
		})
	}

	// Configure Log Level
	lvl, err := logrus.ParseLevel(levelStr)
	if err != nil {
		lvl = logrus.InfoLevel
	}
	globalLogger.SetLevel(lvl)

	return globalLogger
}

// Get returns the global logger instance.
func Get() *logrus.Logger {
	mu.RLock()
	defer mu.RUnlock()
	return globalLogger
}

// Package-level direct logging helpers

func Info(args ...interface{}) {
	Get().Info(args...)
}

func Infof(format string, args ...interface{}) {
	Get().Infof(format, args...)
}

func Warn(args ...interface{}) {
	Get().Warn(args...)
}

func Warnf(format string, args ...interface{}) {
	Get().Warnf(format, args...)
}

func Error(args ...interface{}) {
	Get().Error(args...)
}

func Errorf(format string, args ...interface{}) {
	Get().Errorf(format, args...)
}

func Debug(args ...interface{}) {
	Get().Debug(args...)
}

func Debugf(format string, args ...interface{}) {
	Get().Debugf(format, args...)
}

func Trace(args ...interface{}) {
	Get().Trace(args...)
}

func Tracef(format string, args ...interface{}) {
	Get().Tracef(format, args...)
}

func Fatal(args ...interface{}) {
	Get().Fatal(args...)
}

func Fatalf(format string, args ...interface{}) {
	Get().Fatalf(format, args...)
}

func WithField(key string, value interface{}) *logrus.Entry {
	return Get().WithField(key, value)
}

func WithFields(fields Fields) *logrus.Entry {
	return Get().WithFields(fields)
}

func WithError(err error) *logrus.Entry {
	return Get().WithError(err)
}

func WithContext(ctx context.Context) *logrus.Entry {
	return FromContext(ctx)
}
