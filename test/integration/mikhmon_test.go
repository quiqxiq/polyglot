//go:build integration

package integration

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/quixiq/polyglot/internal/driver/mikrotik"
	mikhmon "github.com/quixiq/polyglot/internal/driver/mikrotik/hotspot"
	mikrotiksystem "github.com/quixiq/polyglot/internal/driver/mikrotik/system"
)

func TestMikhmonIntegration_FullWorkflow(t *testing.T) {
	target := mikrotikTestTarget(t)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	drv, err := mikrotik.NewDriver(ctx, target)
	require.NoError(t, err, "gagal konek ke Mikrotik test")
	defer func() { assert.NoError(t, drv.Close()) }()

	testProfileName := fmt.Sprintf("polyglot-mikhmon-%d", time.Now().UnixNano())
	testVoucherPrefix := fmt.Sprintf("tv%d_", time.Now().UnixNano()%1000000)

	// Cleanup
	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cleanupCancel()
		// Cleanup vouchers
		if printUsers, err := drv.Execute(cleanupCtx, mikhmon.NewPrintUsersCommand("")); err == nil {
			for _, u := range mikhmon.ParseUsers(printUsers) {
				if u.Profile == testProfileName {
					_, _ = drv.Execute(cleanupCtx, mikhmon.NewRemoveUserCommand(u.RosID))
				}
			}
		}
		// Cleanup profile
		if printProfiles, err := drv.Execute(cleanupCtx, mikhmon.NewPrintUserProfilesCommand(testProfileName)); err == nil {
			for _, p := range mikhmon.ParseUserProfiles(printProfiles) {
				_, _ = drv.Execute(cleanupCtx, mikhmon.NewRemoveUserProfileCommand(p.RosID))
			}
		}
		// Cleanup scheduler
		if printSch, err := drv.Execute(cleanupCtx, mikrotiksystem.NewPrintSchedulerCommand(mikhmon.MikhmonExpireMonitorName)); err == nil {
			for _, s := range mikrotiksystem.ParseScheduler(printSch) {
				_, _ = drv.Execute(cleanupCtx, mikrotiksystem.NewSetSchedulerCommand(s.RosID, mikrotiksystem.SystemSchedulerParams{
					Name:     mikhmon.MikhmonExpireMonitorName,
					Disabled: true,
				}))
			}
		}
	})

	// 1. Create Mikhmon Profile with On-Login script
	profileCmd := mikhmon.NewAddMikhmonProfileCommand(mikhmon.MikhmonProfileParams{
		Name:            testProfileName,
		Price:           "5000",
		SellingPrice:    "4000",
		Validity:        "1d",
		ExpireMode:      mikhmon.ExpireModeNotify,
		LockUser:        true,
		EnableRecording: true,
	})
	_, err = drv.Execute(ctx, profileCmd)
	require.NoError(t, err, "gagal create mikhmon profile")

	// Verify profile creation & on-login script
	printProfiles, err := drv.Execute(ctx, mikhmon.NewPrintUserProfilesCommand(testProfileName))
	require.NoError(t, err)
	profiles := mikhmon.ParseUserProfiles(printProfiles)
	require.NotEmpty(t, profiles)
	prof := profiles[0]
	assert.Equal(t, testProfileName, prof.Name)
	assert.Contains(t, prof.OnLogin, ",ntf,5000,1d,4000,,Enable,Disable,")
	t.Logf("Mikhmon profile '%s' berhasil dibuat dengan On-Login script", prof.Name)

	// 2. Generate Batch Mikhmon Vouchers
	batch := mikhmon.NewGenerateVoucherBatchCommands(mikhmon.VoucherGenerateParams{
		Profile:    testProfileName,
		Prefix:     testVoucherPrefix,
		UserLength: 5,
		CharSet:    mikhmon.CharSetUpperNum,
		CommentTag: "IntegrationTest",
	}, 2)

	for _, cmd := range batch.Commands {
		_, err := drv.Execute(ctx, cmd)
		require.NoError(t, err, "gagal add mikhmon voucher")
	}

	// Verify vouchers
	printUsers, err := drv.Execute(ctx, mikhmon.NewPrintUsersCommand(""))
	require.NoError(t, err)
	allUsers := mikhmon.ParseUsers(printUsers)
	var createdUsers []mikhmon.HotspotUser
	for _, u := range allUsers {
		if u.Profile == testProfileName {
			createdUsers = append(createdUsers, u)
		}
	}
	require.Len(t, createdUsers, 2, "harus ada 2 voucher yang terbuat untuk profile ini")
	t.Logf("Voucher 1: %s (comment=%s)", createdUsers[0].Name, createdUsers[0].Comment)
	t.Logf("Voucher 2: %s (comment=%s)", createdUsers[1].Name, createdUsers[1].Comment)

	// Verify comment parsing
	parsedComment, err := mikhmon.ParseMikhmonComment(createdUsers[0].Comment)
	require.NoError(t, err)
	assert.Equal(t, "vc", parsedComment.Type)
	assert.Equal(t, "IntegrationTest", parsedComment.Tag)
	assert.False(t, parsedComment.IsActivated)

	// 3. Verify Inactive Hotspot Users comparison
	activeCmd := mikhmon.NewPrintActiveCommand("")
	activeRes, err := drv.Execute(ctx, activeCmd)
	require.NoError(t, err)
	activeSessions := mikhmon.ParseActiveSessions(activeRes)

	inactiveUsers := mikhmon.FilterInactiveUsers(allUsers, activeSessions)
	require.NotEmpty(t, inactiveUsers, "karena tidak ada sesi aktif, semua user harus terdeteksi inactive")
	t.Logf("Total registered users: %d, active: %d, inactive: %d", len(allUsers), len(activeSessions), len(inactiveUsers))

	// 4. Setup Mikhmon Expire Monitor Scheduler
	// Pre-cleanup if exists
	printSch, err := drv.Execute(ctx, mikrotiksystem.NewPrintSchedulerCommand(mikhmon.MikhmonExpireMonitorName))
	if err == nil {
		schedulers := mikrotiksystem.ParseScheduler(printSch)
		if len(schedulers) > 0 {
			// update existing
			_, err = drv.Execute(ctx, mikhmon.NewUpdateMikhmonExpireMonitorCommand(schedulers[0].RosID, "00:01:00"))
			require.NoError(t, err)
		} else {
			// add new
			_, err = drv.Execute(ctx, mikhmon.NewSetupMikhmonExpireMonitorCommand("00:01:00"))
			require.NoError(t, err)
		}
	}

	// Verify scheduler
	printSchFinal, err := drv.Execute(ctx, mikrotiksystem.NewPrintSchedulerCommand(mikhmon.MikhmonExpireMonitorName))
	require.NoError(t, err)
	finalSchedulers := mikrotiksystem.ParseScheduler(printSchFinal)
	require.NotEmpty(t, finalSchedulers)
	assert.Equal(t, mikhmon.MikhmonExpireMonitorName, finalSchedulers[0].Name)
	assert.Contains(t, finalSchedulers[0].OnEvent, "dateint")
	t.Logf("Mikhmon-Expire-Monitor scheduler berhasil terpasang (interval=%s)", finalSchedulers[0].Interval)
}
