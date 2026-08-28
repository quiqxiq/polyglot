package hotspot

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/quixiq/polyglot/internal/driver/mikrotik/system"
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

func TestClassifyExpireMonitorSchedulers(t *testing.T) {
	t.Run("legacy scheduler installed and enabled", func(t *testing.T) {
		schedulers := []system.SystemScheduler{
			{RosID: "*1", Name: "admin-job", Disabled: false},
			{RosID: "*2", Name: MikhmonExpireMonitorName, Disabled: false},
		}
		status := classifyExpireMonitorSchedulers(schedulers)
		require.True(t, status.IsInstalled)
		require.True(t, status.IsEnabled)
		assert.Equal(t, "*2", status.SchedulerID)
		assert.Equal(t, MikhmonExpireMonitorName, status.SchedulerName)
	})

	t.Run("gateway scheduler installed but disabled", func(t *testing.T) {
		schedulers := []system.SystemScheduler{
			{RosID: "*9", Name: mikhmonExpireSchedulerName, Disabled: true},
		}
		status := classifyExpireMonitorSchedulers(schedulers)
		require.True(t, status.IsInstalled)
		require.False(t, status.IsEnabled, "scheduler disabled → is_enabled=false")
		assert.Equal(t, "*9", status.SchedulerID)
	})

	t.Run("not installed", func(t *testing.T) {
		schedulers := []system.SystemScheduler{
			{RosID: "*1", Name: "backup", Disabled: false},
		}
		status := classifyExpireMonitorSchedulers(schedulers)
		require.False(t, status.IsInstalled)
		assert.Empty(t, status.SchedulerID)
	})

	t.Run("legacy preferred when both forms present", func(t *testing.T) {
		schedulers := []system.SystemScheduler{
			{RosID: "*3", Name: MikhmonExpireMonitorName, Disabled: false},
			{RosID: "*4", Name: mikhmonExpireSchedulerName, Disabled: true},
		}
		status := classifyExpireMonitorSchedulers(schedulers)
		assert.Equal(t, MikhmonExpireMonitorName, status.SchedulerName)
		assert.Equal(t, "*3", status.SchedulerID)
	})
}
