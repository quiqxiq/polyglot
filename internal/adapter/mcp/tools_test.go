package mcp

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/quixiq/polyglot/internal/domain/command"
)

func TestFormatStatus_WithUptimeAndVersion(t *testing.T) {
	result := command.Result{
		Rows: []map[string]string{
			{"uptime": "1d2h", "version": "7.15", "board": "hAP"},
		},
	}
	out := formatStatus("dev1", result)
	assert.Equal(t, "dev1", out.DeviceID)
	assert.Equal(t, "online", out.Status)
	assert.Contains(t, out.Summary, "uptime: 1d2h")
	assert.Contains(t, out.Summary, "version: 7.15")
	assert.Len(t, out.Rows, 1)
}

func TestFormatStatus_WithLastInform(t *testing.T) {
	result := command.Result{
		Rows: []map[string]string{
			{"_lastInform": "2024-01-01T00:00:00Z", "_id": "dev1"},
		},
	}
	out := formatStatus("dev1", result)
	assert.Contains(t, out.Summary, "_lastInform")
}

func TestFormatStatus_NoRows_FallsBackToOutput(t *testing.T) {
	result := command.Result{Output: "reboot ok"}
	out := formatStatus("dev1", result)
	assert.Equal(t, "reboot ok", out.Summary)
	assert.Empty(t, out.Rows)
}

func TestFormatStatus_EmptyResult(t *testing.T) {
	out := formatStatus("dev1", command.Result{})
	assert.Equal(t, "no status data", out.Summary)
	assert.Empty(t, out.Rows)
}

func TestFormatStatus_MultipleRows_UsesFirstForSummary(t *testing.T) {
	result := command.Result{
		Rows: []map[string]string{
			{"uptime": "1d"},
			{"uptime": "2d"},
		},
	}
	out := formatStatus("dev1", result)
	assert.Contains(t, out.Summary, "1d")
	assert.Len(t, out.Rows, 2)
}
