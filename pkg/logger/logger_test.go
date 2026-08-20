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
