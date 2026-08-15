package logger_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/quixiq/polyglot/pkg/logger"
)

func TestLoggerDevelopmentFormatter(t *testing.T) {
	var buf bytes.Buffer
	l := logger.Init("development", "debug", &buf)
	if l == nil {
		t.Fatal("expected logger instance")
	}

	ctx := context.Background()
	ctx = logger.WithRequestID(ctx, "req-12345")
	ctx = logger.WithTenantID(ctx, "tenant-test")

	logger.FromContext(ctx).Info("Test development message")

	output := buf.String()
	if !strings.Contains(output, "Test development message") {
		t.Errorf("expected output to contain message, got: %s", output)
	}
	if !strings.Contains(output, "req-12345") {
		t.Errorf("expected output to contain request_id, got: %s", output)
	}
	if !strings.Contains(output, "tenant-test") {
		t.Errorf("expected output to contain tenant_id, got: %s", output)
	}
}

func TestLoggerProductionJSONFormatter(t *testing.T) {
	var buf bytes.Buffer
	_ = logger.Init("production", "info", &buf)

	logger.WithFields(logger.Fields{
		"action":    "device_reboot",
		"device_id": "dev-999",
	}).Info("Device reboot triggered")

	var parsed map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("failed to parse JSON log output: %v, raw: %s", err, buf.String())
	}

	if parsed["msg"] != "Device reboot triggered" {
		t.Errorf("expected msg field, got: %v", parsed["msg"])
	}
	if parsed["level"] != "info" {
		t.Errorf("expected level 'info', got: %v", parsed["level"])
	}
	if parsed["device_id"] != "dev-999" {
		t.Errorf("expected device_id 'dev-999', got: %v", parsed["device_id"])
	}
}

func TestLoggerDirectHelpers(t *testing.T) {
	var buf bytes.Buffer
	_ = logger.Init("development", "trace", &buf)

	logger.Debugf("Debug count: %d", 42)
	logger.Warn("Warning message")
	logger.WithError(errors.New("connection reset")).Error("Failed to connect")

	output := buf.String()
	if !strings.Contains(output, "Debug count: 42") {
		t.Errorf("expected debug output, got: %s", output)
	}
	if !strings.Contains(output, "Warning message") {
		t.Errorf("expected warn output, got: %s", output)
	}
	if !strings.Contains(output, "connection reset") {
		t.Errorf("expected error output, got: %s", output)
	}
}
