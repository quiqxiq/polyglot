package mikhmon

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildExpireMonitorScript(t *testing.T) {
	script := BuildExpireMonitorScript()

	// Verifikasi skrip mengandung fungsi dateint dan timeint
	assert.Contains(t, script, ":local dateint do={")
	assert.Contains(t, script, ":local timeint do={")
	// Verifikasi query find user dengan regex year
	assert.Contains(t, script, "/ip hotspot user find where comment~\"/$tyear\" || comment~\"/$lyear\"")
	// Verifikasi penanganan Mode N (Notify) dan Mode X (Remove)
	assert.Contains(t, script, "[:pic $comment 21] = \"N\"")
	assert.Contains(t, script, "/ip hotspot user set limit-uptime=1s")
	assert.Contains(t, script, "/ip hotspot user remove")
}

func TestNewSetupMikhmonExpireMonitorCommand(t *testing.T) {
	cmd := NewSetupMikhmonExpireMonitorCommand("00:01:00")

	assert.Equal(t, "/system/scheduler/add", cmd.Raw)
	assert.Equal(t, MikhmonExpireMonitorName, cmd.Args["name"])
	assert.Equal(t, "00:01:00", cmd.Args["interval"])
	assert.NotEmpty(t, cmd.Args["on-event"])
	assert.Equal(t, "no", cmd.Args["disabled"])
}
