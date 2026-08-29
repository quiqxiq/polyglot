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
