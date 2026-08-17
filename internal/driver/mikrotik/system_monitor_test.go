package mikrotik

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/quixiq/polyglot/internal/domain/command"
)

// ─── System health ────────────────────────────────────────────────────────

func TestNewPrintSystemHealthCommand(t *testing.T) {
	cmd := NewPrintSystemHealthCommand()
	assert.Equal(t, "/system/health/print", cmd.Raw)
	assert.False(t, isStreamingCommand(cmd))
}

func TestNewStreamSystemHealthCommand(t *testing.T) {
	t.Run("default interval 1s", func(t *testing.T) {
		cmd := NewStreamSystemHealthCommand("")
		assert.Equal(t, "/system/health/print", cmd.Raw)
		assert.Equal(t, "1s", cmd.Args["interval"])
		assert.True(t, isStreamingCommand(cmd))
	})
	t.Run("custom interval", func(t *testing.T) {
		cmd := NewStreamSystemHealthCommand("2s")
		assert.Equal(t, "2s", cmd.Args["interval"])
		assert.True(t, isStreamingCommand(cmd))
	})
}

func TestParseSystemHealth(t *testing.T) {
	t.Run("row lengkap", func(t *testing.T) {
		result := command.Result{Rows: []map[string]string{
			{
				"voltage": "23.8", "temperature": "52", "cpu-temperature": "57",
				"psu-voltage": "23.8", "psu-current": "0.4", "psu-temperature": "38",
				"fan1-speed": "1912", "fan2-speed": "1931",
			},
		}}
		h := ParseSystemHealth(result)
		assert.Equal(t, "23.8", h.Voltage)
		assert.Equal(t, "52", h.Temperature)
		assert.Equal(t, "57", h.CPUTemperature)
		assert.Equal(t, "0.4", h.PSUCurrent)
		assert.Equal(t, "1912", h.Fan1Speed)
	})
	t.Run("board tanpa sensor — field kosong", func(t *testing.T) {
		result := command.Result{Rows: []map[string]string{{"voltage": "23.8"}}}
		h := ParseSystemHealth(result)
		assert.Equal(t, "23.8", h.Voltage)
		assert.Empty(t, h.Temperature)
		assert.Empty(t, h.Fan1Speed)
	})
	t.Run("result kosong", func(t *testing.T) {
		assert.Empty(t, ParseSystemHealth(command.Result{}))
	})
}

// ─── System routerboard ───────────────────────────────────────────────────

func TestNewStreamSystemRouterboardCommand(t *testing.T) {
	cmd := NewStreamSystemRouterboardCommand("")
	assert.Equal(t, "/system/routerboard/print", cmd.Raw)
	assert.Equal(t, "1s", cmd.Args["interval"])
	assert.True(t, isStreamingCommand(cmd))
}

func TestParseSystemRouterboard(t *testing.T) {
	t.Run("row lengkap", func(t *testing.T) {
		result := command.Result{Rows: []map[string]string{
			{
				"board-name": "hEX PoE", "model": "RB960PGS",
				"serial-number": "ABC123", "firmware-type": "FPGA",
				"factory-firmware": "6.48.1", "current-firmware": "7.15",
				"upgrade-firmware": "7.15.1",
			},
		}}
		rb := ParseSystemRouterboard(result)
		assert.Equal(t, "hEX PoE", rb.BoardName)
		assert.Equal(t, "RB960PGS", rb.Model)
		assert.Equal(t, "ABC123", rb.SerialNumber)
		assert.Equal(t, "7.15", rb.CurrentFirmware)
		assert.Equal(t, "7.15.1", rb.UpgradeFirmware)
	})
	t.Run("result kosong", func(t *testing.T) {
		assert.Empty(t, ParseSystemRouterboard(command.Result{}))
	})
}

// ─── System clock & identity stream ───────────────────────────────────────

func TestNewStreamSystemClockCommand(t *testing.T) {
	cmd := NewStreamSystemClockCommand("")
	assert.Equal(t, "/system/clock/print", cmd.Raw)
	assert.Equal(t, "1s", cmd.Args["interval"])
	assert.True(t, isStreamingCommand(cmd), "interval pada /print harus terdeteksi streaming")
}

func TestNewStreamSystemIdentityCommand(t *testing.T) {
	cmd := NewStreamSystemIdentityCommand("5s")
	assert.Equal(t, "/system/identity/print", cmd.Raw)
	assert.Equal(t, "5s", cmd.Args["interval"])
	assert.True(t, isStreamingCommand(cmd))
}

func TestParseSystemClock(t *testing.T) {
	result := command.Result{Rows: []map[string]string{
		{"time": "14:30:25", "date": "aug/17/2026", "time-zone-name": "Asia/Jakarta", "gmt-offset": "+07:00"},
	}}
	c := ParseSystemClock(result)
	assert.Equal(t, "14:30:25", c.Time)
	assert.Equal(t, "Asia/Jakarta", c.TimeZoneName)
	assert.Equal(t, "+07:00", c.GMTOffset)
}

// ─── Log stream ───────────────────────────────────────────────────────────

func TestNewStreamLogsCommand(t *testing.T) {
	t.Run("tanpa filter", func(t *testing.T) {
		cmd := NewStreamLogsCommand("")
		assert.Equal(t, "/log/print", cmd.Raw)
		assert.True(t, isStreamingCommand(cmd), "follow harus streaming")
	})
	t.Run("dengan filter topik", func(t *testing.T) {
		cmd := NewStreamLogsCommand("hotspot")
		assert.Equal(t, "hotspot", cmd.Args["?topics~"])
		assert.True(t, isStreamingCommand(cmd))
	})
}

func TestParseLogs(t *testing.T) {
	result := command.Result{Rows: []map[string]string{
		{".id": "*1", "time": "14:30:25", "topics": "hotspot,info", "message": "user logged in"},
		{"time": "orphan", "message": "tanpa .id"},
	}}
	logs := ParseLogs(result)
	require.Len(t, logs, 1)
	assert.Equal(t, "*1", logs[0].RosID)
	assert.Equal(t, "hotspot,info", logs[0].Topics)
	assert.Equal(t, "user logged in", logs[0].Message)
}

// ─── Interface ethernet stream ────────────────────────────────────────────

func TestNewStreamInterfacesCommand(t *testing.T) {
	t.Run("full list interval", func(t *testing.T) {
		cmd := NewStreamInterfacesCommand("", "")
		assert.Equal(t, "/interface/ethernet/print", cmd.Raw)
		assert.Equal(t, "1s", cmd.Args["interval"])
		assert.True(t, isStreamingCommand(cmd))
	})
	t.Run("filter nama", func(t *testing.T) {
		cmd := NewStreamInterfacesCommand("ether1", "2s")
		assert.Equal(t, "ether1", cmd.Args["?name"])
		assert.Equal(t, "2s", cmd.Args["interval"])
		assert.True(t, isStreamingCommand(cmd))
	})
}

// ─── Follow streaming commands (real-time events) ─────────────────────────

func TestNewStreamHotspotUsersCommand(t *testing.T) {
	t.Run("tanpa filter", func(t *testing.T) {
		cmd := NewStreamHotspotUsersCommand("")
		assert.Equal(t, "/ip/hotspot/user/print", cmd.Raw)
		assert.Contains(t, cmd.Args, "follow")
		assert.True(t, isStreamingCommand(cmd))
	})
	t.Run("filter profile", func(t *testing.T) {
		cmd := NewStreamHotspotUsersCommand("vip")
		assert.Equal(t, "/ip/hotspot/user/print", cmd.Raw)
		assert.Equal(t, "vip", cmd.Args["?profile"])
		assert.Contains(t, cmd.Args, "follow")
		assert.True(t, isStreamingCommand(cmd))
	})
}

func TestNewStreamHotspotActiveCommand(t *testing.T) {
	t.Run("tanpa filter", func(t *testing.T) {
		cmd := NewStreamHotspotActiveCommand("")
		assert.Equal(t, "/ip/hotspot/active/print", cmd.Raw)
		assert.Contains(t, cmd.Args, "follow")
		assert.True(t, isStreamingCommand(cmd))
	})
	t.Run("filter user", func(t *testing.T) {
		cmd := NewStreamHotspotActiveCommand("budi")
		assert.Equal(t, "/ip/hotspot/active/print", cmd.Raw)
		assert.Equal(t, "budi", cmd.Args["?user"])
		assert.Contains(t, cmd.Args, "follow")
		assert.True(t, isStreamingCommand(cmd))
	})
}

func TestNewStreamPPPoESecretsCommand(t *testing.T) {
	cmd := NewStreamPPPoESecretsCommand("budi")
	assert.Equal(t, "/ppp/secret/print", cmd.Raw)
	assert.Equal(t, "budi", cmd.Args["?name"])
	assert.Contains(t, cmd.Args, "follow")
	assert.True(t, isStreamingCommand(cmd))
}
