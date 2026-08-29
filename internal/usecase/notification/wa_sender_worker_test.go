package notification_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domainNotification "github.com/quixiq/polyglot/internal/domain/notification"
	"github.com/quixiq/polyglot/internal/port/mocktest"
	uc "github.com/quixiq/polyglot/internal/usecase/notification"
)

func newWorker(t *testing.T, notif *mocktest.FakeNotificationRepo, sender *mocktest.FakeNotificationSender) *uc.WASenderWorker {
	t.Helper()
	settings := mocktest.NewFakeSettingReader(map[string]string{"isp.wa_send_max_retry": "3"})
	return uc.NewWASenderWorker(notif, sender, settings)
}

func queue(t *testing.T, notif *mocktest.FakeNotificationRepo, id, phone string) {
	t.Helper()
	require.NoError(t, notif.Queue(context.Background(), domainNotification.WANotification{
		ID: id, TenantID: "tenant-default", RecipientPhone: phone,
		MessageType: "BILL_REMINDER", MessageContent: "halo " + id,
		Status: domainNotification.StatusQueued, // DB prod punya default; fake tidak
	}))
}

func TestWaSender_SendsPending_MarksSent(t *testing.T) {
	notif := mocktest.NewFakeNotificationRepo()
	sender := &mocktest.FakeNotificationSender{}
	queue(t, notif, "wa-1", "08111")
	queue(t, notif, "wa-2", "+62 812 2222 3333") // format berantakan → dinormalisasi

	res, err := newWorker(t, notif, sender).Run(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 2, res.Sent)
	assert.Len(t, sender.Messages, 2)
	// Worker meneruskan nomor mentah; NORMALISASI ke 62... adalah tanggung
	// jawab SenderAdapter (lihat whatsappadapter). Di sini verifikasi
	// passthrough apa adanya.
	assert.Equal(t, "+62 812 2222 3333", sender.Messages[1].Phone)

	pending, _ := notif.Pending(context.Background(), 10)
	assert.Empty(t, pending)
}

func TestWaSender_Failure_IncrementsAttempt_ThenGivesUp(t *testing.T) {
	notif := mocktest.NewFakeNotificationRepo()
	sender := &mocktest.FakeNotificationSender{Err: errors.New("gateway down")}
	queue(t, notif, "wa-f", "081234567890")

	worker := newWorker(t, notif, sender)
	for i := 1; i <= 3; i++ {
		res, err := worker.Run(context.Background())
		require.NoError(t, err)
		if i < 3 {
			assert.Equal(t, 0, res.GaveUp)
		} else {
			assert.Equal(t, 1, res.GaveUp, "siklus ke-3: menyerah permanen")
		}
	}
	// Siklus keempat tidak mengirim apa pun.
	res, err := worker.Run(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 0, res.Failed)
	assert.Equal(t, 0, res.GaveUp)
}

func TestWaSender_NoSession_DoesNotBurnAttempts(t *testing.T) {
	notif := mocktest.NewFakeNotificationRepo()
	sender := &mocktest.FakeNotificationSender{Err: domainNotification.ErrNoWASession}
	queue(t, notif, "wa-n", "081234567890")

	res, err := newWorker(t, notif, sender).Run(context.Background())
	require.NoError(t, err)
	assert.True(t, res.NoSession)

	all := notif.AllNotifications()
	require.NotEmpty(t, all)
	assert.Equal(t, 0, all[0].Attempts, "infrastruktur mati tak boleh membakar attempts")
}
