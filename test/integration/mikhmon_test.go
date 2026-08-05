//go:build integration

package integration

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/quixiq/polyglot/internal/driver/mikrotik"
	"github.com/quixiq/polyglot/internal/driver/mikrotik/mikhmon"
)

func TestMikhmonIntegration_FullWorkflow(t *testing.T) {
	drv := newTestDriver(t)
	ctx := context.Background()

	const testProfileName = "polyglot-mikhmon-profile"
	const testVoucherPrefix = "testvc_"

	// Cleanup
	t.Cleanup(func() {
		// Cleanup vouchers
		if printUsers, err := drv.Execute(context.Background(), mikrotik.NewPrintHotspotUsersCommand("")); err == nil {
			for _, u := range mikrotik.ParseHotspotUsers(printUsers) {
				if u.Profile == testProfileName {
					_, _ = drv.Execute(context.Background(), mikrotik.NewRemoveHotspotUserCommand(u.RosID))
				}
			}
		}
		// Cleanup profile
		if printProfiles, err := drv.Execute(context.Background(), mikrotik.NewPrintHotspotUserProfilesCommand(testProfileName)); err == nil {
			for _, p := range mikrotik.ParseHotspotUserProfiles(printProfiles) {
				_, _ = drv.Execute(context.Background(), mikrotik.NewRemoveHotspotUserProfileCommand(p.RosID))
			}
		}
		// Cleanup scheduler
		if printSch, err := drv.Execute(context.Background(), mikrotik.NewPrintSystemSchedulersCommand(mikhmon.MikhmonExpireMonitorName)); err == nil {
			for _, s := range mikrotik.ParseSystemSchedulers(printSch) {
				_, _ = drv.Execute(context.Background(), mikrotik.NewSetSystemSchedulerCommand(s.RosID, mikrotik.SystemSchedulerParams{
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
	_, err := drv.Execute(ctx, profileCmd)
	require.NoError(t, err, "gagal create mikhmon profile")

	// Verify profile creation & on-login script
	printProfiles, err := drv.Execute(ctx, mikrotik.NewPrintHotspotUserProfilesCommand(testProfileName))
	require.NoError(t, err)
	profiles := mikrotik.ParseHotspotUserProfiles(printProfiles)
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
	printUsers, err := drv.Execute(ctx, mikrotik.NewPrintHotspotUsersCommand(""))
	require.NoError(t, err)
	allUsers := mikrotik.ParseHotspotUsers(printUsers)
	var createdUsers []mikrotik.HotspotUser
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

	// 3. Setup Mikhmon Expire Monitor Scheduler
	// Pre-cleanup if exists
	printSch, err := drv.Execute(ctx, mikrotik.NewPrintSystemSchedulersCommand(mikhmon.MikhmonExpireMonitorName))
	if err == nil {
		schedulers := mikrotik.ParseSystemSchedulers(printSch)
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
	printSchFinal, err := drv.Execute(ctx, mikrotik.NewPrintSystemSchedulersCommand(mikhmon.MikhmonExpireMonitorName))
	require.NoError(t, err)
	finalSchedulers := mikrotik.ParseSystemSchedulers(printSchFinal)
	require.NotEmpty(t, finalSchedulers)
	assert.Equal(t, mikhmon.MikhmonExpireMonitorName, finalSchedulers[0].Name)
	assert.Contains(t, finalSchedulers[0].OnEvent, "dateint")
	t.Logf("Mikhmon-Expire-Monitor scheduler berhasil terpasang (interval=%s)", finalSchedulers[0].Interval)
}
