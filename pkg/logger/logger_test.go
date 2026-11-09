package logger_test

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/quixiq/polyglot/pkg/logger"
	"github.com/sirupsen/logrus"
)

func TestLoggerInitAndLevels(t *testing.T) {
	var buf bytes.Buffer
	logger.Init("debug", "development", &buf)

	logger.Debug("debug message")
	logger.Info("info message")
	logger.Warn("warn message")
	logger.Error("error message")

	output := buf.String()
	if !strings.Contains(output, "debug message") {
		t.Errorf("expected debug message in log, got: %s", output)
	}
	if !strings.Contains(output, "info message") {
		t.Errorf("expected info message in log, got: %s", output)
	}
	if !strings.Contains(output, "warn message") {
		t.Errorf("expected warn message in log, got: %s", output)
	}
	if !strings.Contains(output, "error message") {
		t.Errorf("expected error message in log, got: %s", output)
	}
}

func TestLoggerWithComponentAndFields(t *testing.T) {
	var buf bytes.Buffer
	logger.Init("info", "production", &buf)

	logger.WithComponent("DeviceService").WithFields(logrus.Fields{
		"device_id": "dev-123",
		"status":    "online",
	}).Info("device status updated")

	output := buf.String()
	if !strings.Contains(output, `"component":"DeviceService"`) {
		t.Errorf("expected component in JSON log, got: %s", output)
	}
	if !strings.Contains(output, `"device_id":"dev-123"`) {
		t.Errorf("expected device_id in JSON log, got: %s", output)
	}
}

func TestSensitiveFieldsAreRedacted(t *testing.T) {
	var buf bytes.Buffer
	logger.Init("info", "production", &buf)
	t.Cleanup(func() { logger.Init("info", "development") })

	logger.WithComponent("WhatsAppDriver").WithFields(logrus.Fields{
		"token":           "access-token",
		"jid":             "6281234567890@s.whatsapp.net",
		"phone_number":    "+6281234567890",
		"password":        "router-password",
		"request_payload": `{"message":"private"}`,
	}).Info("test log")
	logger.WithComponent("WhatsAppDriver").WithField("error_payload", "sensitive-body").Info("payload test")

	output := buf.String()
	for _, secret := range []string{
		"access-token",
		"6281234567890@s.whatsapp.net",
		"+6281234567890",
		"router-password",
		`{"message":"private"}`,
		"sensitive-body",
	} {
		if strings.Contains(output, secret) {
			t.Errorf("sensitive value %q found in log output %q", secret, output)
		}
	}
	if !strings.Contains(output, "[REDACTED]") {
		t.Errorf("redacted log = %q, want redaction marker", output)
	}
}

func TestLoggerWithErrorAndContext(t *testing.T) {
	var buf bytes.Buffer
	logger.Init("info", "development", &buf)

	ctx := context.Background()
	testErr := errors.New("connection timed out")

	logger.WithContext(ctx).WithError(testErr).Error("failed to connect to host")

	output := buf.String()
	if !strings.Contains(output, "connection timed out") {
		t.Errorf("expected error in log output, got: %s", output)
	}
}

func TestLoggerCorrelationContext(t *testing.T) {
	ctx := context.Background()
	ctx = logger.ContextWithRequestID(ctx, "req-xyz-123")
	ctx = logger.ContextWithUserID(ctx, "user-456")
	ctx = logger.ContextWithDeviceID(ctx, "dev-789")
	ctx = logger.ContextWithOperation(ctx, "provision_hotspot")

	if got := logger.RequestIDFromContext(ctx); got != "req-xyz-123" {
		t.Errorf("RequestID = %q, want req-xyz-123", got)
	}
	if got := logger.UserIDFromContext(ctx); got != "user-456" {
		t.Errorf("UserID = %q, want user-456", got)
	}
	if got := logger.DeviceIDFromContext(ctx); got != "dev-789" {
		t.Errorf("DeviceID = %q, want dev-789", got)
	}
	if got := logger.OperationFromContext(ctx); got != "provision_hotspot" {
		t.Errorf("Operation = %q, want provision_hotspot", got)
	}

	corr := logger.CorrelationFromContext(ctx)
	if corr.RequestID != "req-xyz-123" || corr.UserID != "user-456" || corr.DeviceID != "dev-789" || corr.Operation != "provision_hotspot" {
		t.Errorf("Correlation = %+v, unexpected fields", corr)
	}

	var buf bytes.Buffer
	logger.Init("info", "production", &buf)
	t.Cleanup(func() { logger.Init("info", "development") })

	logger.WithContext(ctx).Info("operation completed successfully")

	out := buf.String()
	for _, expected := range []string{
		`"request_id":"req-xyz-123"`,
		`"user_id":"user-456"`,
		`"device_id":"dev-789"`,
		`"operation":"provision_hotspot"`,
	} {
		if !strings.Contains(out, expected) {
			t.Errorf("expected %s in JSON output, got: %s", expected, out)
		}
	}
}

func TestLoggerWithCorrelationAndOperation(t *testing.T) {
	var buf bytes.Buffer
	logger.Init("info", "production", &buf)
	t.Cleanup(func() { logger.Init("info", "development") })

	logger.WithCorrelation(logger.Correlation{
		RequestID: "req-corr-1",
		UserID:    "user-corr-2",
		DeviceID:  "dev-corr-3",
		Operation: "test_corr",
	}).Info("correlated event")

	out := buf.String()
	if !strings.Contains(out, `"request_id":"req-corr-1"`) ||
		!strings.Contains(out, `"user_id":"user-corr-2"`) ||
		!strings.Contains(out, `"device_id":"dev-corr-3"`) ||
		!strings.Contains(out, `"operation":"test_corr"`) {
		t.Errorf("expected correlation fields in output, got: %s", out)
	}

	buf.Reset()
	logger.WithOperation("custom_op").Info("single operation")
	if !strings.Contains(buf.String(), `"operation":"custom_op"`) {
		t.Errorf("expected operation in output, got: %s", buf.String())
	}
}

func TestLoggerInitConcurrency(t *testing.T) {
	done := make(chan struct{})
	for i := 0; i < 10; i++ {
		go func(idx int) {
			var buf bytes.Buffer
			env := "development"
			if idx%2 == 0 {
				env = "production"
			}
			logger.Init("debug", env, &buf)
			logger.WithComponent("ConcurTest").WithField("worker", idx).Debug("concurrent log")
			done <- struct{}{}
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}
