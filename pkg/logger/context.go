package logger

import (
	"context"

	"github.com/sirupsen/logrus"
)

type contextKey string

const (
	requestIDKey contextKey = "request_id"
	tenantIDKey  contextKey = "tenant_id"
	deviceIDKey  contextKey = "device_id"
)

// WithRequestID attaches a request ID to context.
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey, requestID)
}

// WithTenantID attaches a tenant ID to context.
func WithTenantID(ctx context.Context, tenantID string) context.Context {
	return context.WithValue(ctx, tenantIDKey, tenantID)
}

// WithDeviceID attaches a device ID to context.
func WithDeviceID(ctx context.Context, deviceID string) context.Context {
	return context.WithValue(ctx, deviceIDKey, deviceID)
}

// FromContext extracts contextual identifiers (request_id, tenant_id, device_id)
// and returns a pre-configured *logrus.Entry.
func FromContext(ctx context.Context) *logrus.Entry {
	entry := Get().WithContext(ctx)

	if ctx == nil {
		return entry
	}

	fields := logrus.Fields{}
	if reqID, ok := ctx.Value(requestIDKey).(string); ok && reqID != "" {
		fields["request_id"] = reqID
	}
	if tenantID, ok := ctx.Value(tenantIDKey).(string); ok && tenantID != "" {
		fields["tenant_id"] = tenantID
	}
	if devID, ok := ctx.Value(deviceIDKey).(string); ok && devID != "" {
		fields["device_id"] = devID
	}

	if len(fields) > 0 {
		return entry.WithFields(fields)
	}

	return entry
}

// InfoContext logs at Info level with context fields.
func InfoContext(ctx context.Context, msg string) {
	FromContext(ctx).Info(msg)
}

// ErrorContext logs at Error level with context fields.
func ErrorContext(ctx context.Context, msg string) {
	FromContext(ctx).Error(msg)
}

// WarnContext logs at Warn level with context fields.
func WarnContext(ctx context.Context, msg string) {
	FromContext(ctx).Warn(msg)
}

// DebugContext logs at Debug level with context fields.
func DebugContext(ctx context.Context, msg string) {
	FromContext(ctx).Debug(msg)
}
