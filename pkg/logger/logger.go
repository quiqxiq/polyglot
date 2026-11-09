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
	loggerMu      sync.RWMutex
)

// Fields is an alias for structured log fields.
type Fields = logrus.Fields

type correlationContextKey struct{}

// Correlation holds structured correlation identifiers for request lifecycle tracking.
type Correlation struct {
	RequestID string
	UserID    string
	DeviceID  string
	Operation string
}

func init() {
	once.Do(func() {
		defaultLogger = logrus.New()
		defaultLogger.AddHook(redactionHook{})
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
	loggerMu.Lock()
	defer loggerMu.Unlock()

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
	loggerMu.RLock()
	defer loggerMu.RUnlock()
	return defaultLogger
}

// WithComponent creates an Entry tagged with a component name.
func WithComponent(component string) *logrus.Entry {
	return defaultLogger.WithField("component", component)
}

// WithField adds a single field to the log entry.
func WithField(key string, value any) *logrus.Entry {
	return defaultLogger.WithField(key, redactField(key, value))
}

// WithFields adds multiple fields to the log entry.
func WithFields(fields logrus.Fields) *logrus.Entry {
	redacted := make(logrus.Fields, len(fields))
	for key, value := range fields {
		redacted[key] = redactField(key, value)
	}
	return defaultLogger.WithFields(redacted)
}

// ContextWithRequestID returns a copy of ctx with request ID attached.
func ContextWithRequestID(ctx context.Context, requestID string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	corr := correlationFromCtx(ctx)
	corr.RequestID = requestID
	return context.WithValue(ctx, correlationContextKey{}, corr)
}

// RequestIDFromContext extracts the request ID from ctx, returning an empty string if absent.
func RequestIDFromContext(ctx context.Context) string {
	return correlationFromCtx(ctx).RequestID
}

// ContextWithUserID returns a copy of ctx with user ID attached.
func ContextWithUserID(ctx context.Context, userID string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	corr := correlationFromCtx(ctx)
	corr.UserID = userID
	return context.WithValue(ctx, correlationContextKey{}, corr)
}

// UserIDFromContext extracts the user ID from ctx, returning an empty string if absent.
func UserIDFromContext(ctx context.Context) string {
	return correlationFromCtx(ctx).UserID
}

// ContextWithDeviceID returns a copy of ctx with device ID attached.
func ContextWithDeviceID(ctx context.Context, deviceID string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	corr := correlationFromCtx(ctx)
	corr.DeviceID = deviceID
	return context.WithValue(ctx, correlationContextKey{}, corr)
}

// DeviceIDFromContext extracts the device ID from ctx, returning an empty string if absent.
func DeviceIDFromContext(ctx context.Context) string {
	return correlationFromCtx(ctx).DeviceID
}

// ContextWithOperation returns a copy of ctx with operation name attached.
func ContextWithOperation(ctx context.Context, operation string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	corr := correlationFromCtx(ctx)
	corr.Operation = operation
	return context.WithValue(ctx, correlationContextKey{}, corr)
}

// OperationFromContext extracts the operation from ctx, returning an empty string if absent.
func OperationFromContext(ctx context.Context) string {
	return correlationFromCtx(ctx).Operation
}

// ContextWithCorrelation returns a copy of ctx with all non-empty correlation fields attached.
func ContextWithCorrelation(ctx context.Context, corr Correlation) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	existing := correlationFromCtx(ctx)
	if corr.RequestID != "" {
		existing.RequestID = corr.RequestID
	}
	if corr.UserID != "" {
		existing.UserID = corr.UserID
	}
	if corr.DeviceID != "" {
		existing.DeviceID = corr.DeviceID
	}
	if corr.Operation != "" {
		existing.Operation = corr.Operation
	}
	return context.WithValue(ctx, correlationContextKey{}, existing)
}

// CorrelationFromContext extracts correlation fields from ctx.
func CorrelationFromContext(ctx context.Context) Correlation {
	return correlationFromCtx(ctx)
}

func correlationFromCtx(ctx context.Context) Correlation {
	if ctx == nil {
		return Correlation{}
	}
	val, ok := ctx.Value(correlationContextKey{}).(Correlation)
	if !ok {
		return Correlation{}
	}
	return val
}

// WithContext returns a log entry with context attached and automatically populates
// any non-empty correlation fields (request_id, user_id, device_id, operation) present in ctx.
func WithContext(ctx context.Context) *logrus.Entry {
	entry := defaultLogger.WithContext(ctx)
	if ctx == nil {
		return entry
	}
	corr := correlationFromCtx(ctx)
	fields := logrus.Fields{}
	if corr.RequestID != "" {
		fields["request_id"] = corr.RequestID
	}
	if corr.UserID != "" {
		fields["user_id"] = corr.UserID
	}
	if corr.DeviceID != "" {
		fields["device_id"] = corr.DeviceID
	}
	if corr.Operation != "" {
		fields["operation"] = corr.Operation
	}
	if len(fields) > 0 {
		return entry.WithFields(fields)
	}
	return entry
}

// WithOperation returns an entry tagged with an operation field.
func WithOperation(operation string) *logrus.Entry {
	return defaultLogger.WithField("operation", operation)
}

// WithCorrelation returns an entry tagged with all non-empty correlation fields.
func WithCorrelation(corr Correlation) *logrus.Entry {
	fields := logrus.Fields{}
	if corr.RequestID != "" {
		fields["request_id"] = corr.RequestID
	}
	if corr.UserID != "" {
		fields["user_id"] = corr.UserID
	}
	if corr.DeviceID != "" {
		fields["device_id"] = corr.DeviceID
	}
	if corr.Operation != "" {
		fields["operation"] = corr.Operation
	}
	return defaultLogger.WithFields(fields)
}

// WithError returns a log entry with an error attached.
func WithError(err error) *logrus.Entry {
	if err == nil {
		return defaultLogger.WithFields(nil)
	}
	return defaultLogger.WithError(redactError(err))
}

type redactedError struct{ message string }

func (e redactedError) Error() string { return e.message }

func (e redactedError) Unwrap() error { return nil }

func redactError(err error) error {
	message := err.Error()
	for _, marker := range []string{"token=", "password=", "secret=", "api_key=", "phone=", "jid="} {
		if strings.Contains(strings.ToLower(message), marker) {
			return redactedError{message: "[REDACTED]"}
		}
	}
	return err
}

type redactionHook struct{}

func (redactionHook) Levels() []logrus.Level { return logrus.AllLevels }

func (redactionHook) Fire(entry *logrus.Entry) error {
	for key, value := range entry.Data {
		entry.Data[key] = redactField(key, value)
	}
	if errValue, ok := entry.Data[logrus.ErrorKey].(error); ok {
		entry.Data[logrus.ErrorKey] = redactError(errValue)
	}
	return nil
}

func redactField(key string, value any) any {
	key = strings.ToLower(strings.ReplaceAll(key, "-", "_"))
	for _, sensitive := range []string{
		"password", "token", "secret", "api_key", "private_key", "cookie",
		"authorization", "payload", "jid", "phone",
	} {
		if strings.Contains(key, sensitive) {
			return "[REDACTED]"
		}
	}
	return value
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
