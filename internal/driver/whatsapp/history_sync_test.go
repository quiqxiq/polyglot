package whatsapp

import (
	"context"
	"sync"
	"testing"
	"time"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waCommon"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/proto/waHistorySync"
	"go.mau.fi/whatsmeow/proto/waWeb"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	"google.golang.org/protobuf/proto"

	"github.com/quixiq/polyglot/internal/domain/bot"
	"github.com/quixiq/polyglot/internal/port"
)

// fakeHistoryRepo records mirrored chats/messages/unread untuk pengujian
// history sync mirror tanpa database nyata.
type fakeHistoryRepo struct {
	mu            sync.Mutex
	chats         map[string]*bot.WAChat
	messages      map[string][]*bot.WAMessage
	unread        map[string]uint32
	statusUpdates []statusUpdate
	batchCalls    int
	singleCalls   int
}

// statusUpdate merekam satu panggilan MarkMessagesStatus untuk pengujian.
type statusUpdate struct {
	chatJID    string
	messageIDs []string
	status     string
}

func newFakeHistoryRepo() *fakeHistoryRepo {
	return &fakeHistoryRepo{
		chats:    make(map[string]*bot.WAChat),
		messages: make(map[string][]*bot.WAMessage),
		unread:   make(map[string]uint32),
	}
}

func (f *fakeHistoryRepo) UpsertChat(_ context.Context, c *bot.WAChat) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.chats[c.ChatJID] = c
	return nil
}

func (f *fakeHistoryRepo) UpsertMessage(_ context.Context, m *bot.WAMessage) (bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.singleCalls++
	f.messages[m.ChatJID] = append(f.messages[m.ChatJID], m)
	return true, nil
}

func (f *fakeHistoryRepo) UpsertMessagesBatch(_ context.Context, msgs []*bot.WAMessage) (int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.batchCalls++
	for _, m := range msgs {
		f.messages[m.ChatJID] = append(f.messages[m.ChatJID], m)
	}
	return len(msgs), nil
}

func (f *fakeHistoryRepo) IncrementUnread(_ context.Context, _ uint, _ string) error { return nil }
func (f *fakeHistoryRepo) MarkChatRead(_ context.Context, _ uint, _ string) error          { return nil }
func (f *fakeHistoryRepo) SetChatUnread(_ context.Context, _ uint, chatJID string, count uint32) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.unread[chatJID] = count
	return nil
}
func (f *fakeHistoryRepo) ListChats(_ context.Context, _ uint, _, _ int, _ string) ([]bot.WAChat, error) {
	return nil, nil
}
func (f *fakeHistoryRepo) ListChatMessages(_ context.Context, _ uint, _ string, _, _ int) ([]bot.WAMessage, error) {
	return nil, nil
}
func (f *fakeHistoryRepo) SetChatBotEnabled(_ context.Context, _ uint, _ string, _ bool) error { return nil }
func (f *fakeHistoryRepo) IsChatBotEnabled(_ context.Context, _ uint, _ string) (bool, error)  { return true, nil }
func (f *fakeHistoryRepo) MarkMessagesStatus(_ context.Context, _ uint, chatJID string, messageIDs []string, status string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.statusUpdates = append(f.statusUpdates, statusUpdate{chatJID: chatJID, messageIDs: messageIDs, status: status})
	return nil
}

func (f *fakeHistoryRepo) MergeChatLID(_ context.Context, _ uint, lidJID, pnJID string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	// Pindahkan pesan @lid ke PN, hapus baris chat @lid.
	if msgs := f.messages[lidJID]; msgs != nil {
		f.messages[pnJID] = append(f.messages[pnJID], msgs...)
		delete(f.messages, lidJID)
	}
	delete(f.chats, lidJID)
	return nil
}

var _ port.ChatRepository = (*fakeHistoryRepo)(nil)

func TestHandleHistorySyncMirrorsConversations(t *testing.T) {
	fake := newFakeHistoryRepo()
	c := &Client{SessionID: 7, chatRepo: fake, waClient: &whatsmeow.Client{}}

	timestamp := uint64(1700000000)

	// 1:1 chat dengan 2 pesan (masuk + keluar) dan unread.
	direct := &waHistorySync.Conversation{
		ID:          proto.String("628123456789@s.whatsapp.net"),
		Name:        proto.String("Budi"),
		UnreadCount: proto.Uint32(2),
		Messages: []*waHistorySync.HistorySyncMsg{
			{
				Message: &waWeb.WebMessageInfo{
					Key: &waCommon.MessageKey{
						RemoteJID: proto.String("628123456789@s.whatsapp.net"),
						FromMe:    proto.Bool(false),
						ID:        proto.String("id-in-1"),
					},
					Message:          &waE2E.Message{Conversation: proto.String("halo")},
					MessageTimestamp: proto.Uint64(timestamp),
					PushName:         proto.String("Budi"),
				},
			},
			{
				Message: &waWeb.WebMessageInfo{
					Key: &waCommon.MessageKey{
						RemoteJID: proto.String("628123456789@s.whatsapp.net"),
						FromMe:    proto.Bool(true),
						ID:        proto.String("id-out-1"),
					},
					Message:          &waE2E.Message{ExtendedTextMessage: &waE2E.ExtendedTextMessage{Text: proto.String("baik, nanti saya cek")}},
					MessageTimestamp: proto.Uint64(timestamp + 10),
				},
			},
		},
	}

	// Grup dengan 1 pesan dari anggota (sender ada di level WebMessageInfo).
	group := &waHistorySync.Conversation{
		ID:   proto.String("1234567890-1612345678@g.us"),
		Name: proto.String("NetOps Team"),
		Messages: []*waHistorySync.HistorySyncMsg{
			{
				Message: &waWeb.WebMessageInfo{
					Key: &waCommon.MessageKey{
						RemoteJID: proto.String("1234567890-1612345678@g.us"),
						FromMe:    proto.Bool(false),
						ID:        proto.String("id-grp-1"),
					},
					Message:          &waE2E.Message{ImageMessage: &waE2E.ImageMessage{Caption: proto.String("foto topologi")}},
					MessageTimestamp: proto.Uint64(timestamp + 20),
					Participant:      proto.String("628999@s.whatsapp.net"),
					PushName:         proto.String("Citra"),
				},
			},
		},
	}

	evt := &events.HistorySync{Data: &waHistorySync.HistorySync{
		SyncType:      waHistorySync.HistorySync_INITIAL_BOOTSTRAP.Enum(),
		Conversations: []*waHistorySync.Conversation{direct, group},
	}}
	c.handleHistorySync(evt)

	// handleHistorySync berjalan di goroutine — tunggu sampai 2 chat tercatat.
	waitFor(t, 2*time.Second, func() bool {
		fake.mu.Lock()
		defer fake.mu.Unlock()
		return len(fake.chats) == 2 &&
			len(fake.messages["628123456789@s.whatsapp.net"]) == 2 &&
			len(fake.messages["1234567890-1612345678@g.us"]) == 1
	})

	fake.mu.Lock()
	defer fake.mu.Unlock()

	// Sinkronisasi harus memakai batch insert, bukan satu query per pesan.
	if fake.batchCalls != 2 {
		t.Errorf("UpsertMessagesBatch dipanggil %d kali, want 2 (satu per conversation)", fake.batchCalls)
	}
	if fake.singleCalls != 0 {
		t.Errorf("UpsertMessage (per-pesan) dipanggil %d kali, want 0 untuk history sync", fake.singleCalls)
	}

	// Chat 1:1 — nama, preview dari pesan terakhir (keluar), unread 2.
	dc, ok := fake.chats["628123456789@s.whatsapp.net"]
	if !ok {
		t.Fatal("chat 1:1 tidak dimirror")
	}
	if dc.DisplayName != "Budi" || dc.IsGroup {
		t.Errorf("chat 1:1: name=%q isGroup=%v, want Budi/false", dc.DisplayName, dc.IsGroup)
	}
	if dc.LastMessagePreview != "baik, nanti saya cek" {
		t.Errorf("chat 1:1 preview = %q", dc.LastMessagePreview)
	}
	if got := fake.unread["628123456789@s.whatsapp.net"]; got != 2 {
		t.Errorf("chat 1:1 unread = %d, want 2", got)
	}

	// Chat grup — nama grup, isGroup true, preview media.
	gc, ok := fake.chats["1234567890-1612345678@g.us"]
	if !ok {
		t.Fatal("chat grup tidak dimirror")
	}
	if gc.DisplayName != "NetOps Team" || !gc.IsGroup {
		t.Errorf("chat grup: name=%q isGroup=%v, want NetOps Team/true", gc.DisplayName, gc.IsGroup)
	}
	// Pesan terakhir grup adalah gambar ber-caption → preview memakai caption.
	if gc.LastMessagePreview != "foto topologi" {
		t.Errorf("chat grup preview = %q, want caption", gc.LastMessagePreview)
	}

	// Pesan: 1:1 masuk (sender = chat JID), 1:1 keluar (from me), grup (sender dari participant).
	if len(fake.messages["628123456789@s.whatsapp.net"]) != 2 {
		t.Fatalf("jumlah pesan 1:1 = %d, want 2", len(fake.messages["628123456789@s.whatsapp.net"]))
	}
	incoming := fake.messages["628123456789@s.whatsapp.net"][0]
	if incoming.Content != "halo" || incoming.IsFromMe || incoming.SenderJID != "628123456789@s.whatsapp.net" {
		t.Errorf("pesan masuk: content=%q fromMe=%v sender=%q", incoming.Content, incoming.IsFromMe, incoming.SenderJID)
	}
	outgoing := fake.messages["628123456789@s.whatsapp.net"][1]
	if !outgoing.IsFromMe {
		t.Error("pesan keluar harus is_from_me=true")
	}
	grpMsg := fake.messages["1234567890-1612345678@g.us"][0]
	if grpMsg.Content != "foto topologi" || grpMsg.MediaType != "image" || grpMsg.SenderJID != "628999@s.whatsapp.net" {
		t.Errorf("pesan grup: content=%q media=%q sender=%q", grpMsg.Content, grpMsg.MediaType, grpMsg.SenderJID)
	}
}

func TestHandleHistorySyncPushNames(t *testing.T) {
	fake := newFakeHistoryRepo()
	c := &Client{SessionID: 7, chatRepo: fake, waClient: &whatsmeow.Client{}}

	evt := &events.HistorySync{Data: &waHistorySync.HistorySync{
		SyncType: waHistorySync.HistorySync_PUSH_NAME.Enum(),
		Pushnames: []*waHistorySync.Pushname{
			{ID: proto.String("628111@s.whatsapp.net"), Pushname: proto.String("Andi")},
		},
	}}
	c.handleHistorySync(evt)

	waitFor(t, 2*time.Second, func() bool {
		fake.mu.Lock()
		defer fake.mu.Unlock()
		return len(fake.chats) == 1
	})

	fake.mu.Lock()
	defer fake.mu.Unlock()
	chat, ok := fake.chats["628111@s.whatsapp.net"]
	if !ok {
		t.Fatal("chat dari push name tidak dibuat")
	}
	if chat.DisplayName != "Andi" {
		t.Errorf("display name = %q, want Andi", chat.DisplayName)
	}
}

// waitFor polls cond sampai true atau timeout — dipakai karena handler
// history sync berjalan asinkron (goroutine).
func waitFor(t *testing.T, timeout time.Duration, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("condition tidak terpenuhi sebelum timeout")
}

// TestHandleReceiptEventMarksStatus memverifikasi bahwa events.Receipt
// diterjemahkan ke status WhatsApp (delivered ✓✓ / read ✓✓ biru) dan diteruskan
// ke ChatRepository untuk pesan keluar. Ini melengkapi 4-status pesan:
// sent (✓) di-set saat kirim, delivered/read di-set saat receipt tiba.
func TestHandleReceiptEventMarksStatus(t *testing.T) {
	repo := newFakeHistoryRepo()
	c := &Client{SessionID: 7, chatRepo: repo, waClient: &whatsmeow.Client{}}

	delivered := &events.Receipt{
		MessageSource: types.MessageSource{
			Chat:   types.NewJID("628123456789", types.DefaultUserServer),
			Sender: types.NewJID("628123456789", types.DefaultUserServer),
		},
		MessageIDs: []types.MessageID{"id-delivered-1"},
		Type:       types.ReceiptTypeDelivered,
		Timestamp:  time.Now(),
	}
	c.handleReceiptEvent(delivered)

	read := &events.Receipt{
		MessageSource: types.MessageSource{
			Chat:   types.NewJID("628123456789", types.DefaultUserServer),
			Sender: types.NewJID("628123456789", types.DefaultUserServer),
		},
		MessageIDs: []types.MessageID{"id-read-1", "id-read-2"},
		Type:       types.ReceiptTypeRead,
		Timestamp:  time.Now(),
	}
	c.handleReceiptEvent(read)

	repo.mu.Lock()
	defer repo.mu.Unlock()
	if len(repo.statusUpdates) != 2 {
		t.Fatalf("statusUpdates = %d, want 2", len(repo.statusUpdates))
	}
	if repo.statusUpdates[0].status != "delivered" || len(repo.statusUpdates[0].messageIDs) != 1 {
		t.Errorf("update[0] = %+v, want delivered with 1 message id", repo.statusUpdates[0])
	}
	if repo.statusUpdates[0].chatJID != "628123456789@s.whatsapp.net" {
		t.Errorf("update[0].chatJID = %q, want 628123456789@s.whatsapp.net", repo.statusUpdates[0].chatJID)
	}
	if repo.statusUpdates[1].status != "read" || len(repo.statusUpdates[1].messageIDs) != 2 {
		t.Errorf("update[1] = %+v, want read with 2 message ids", repo.statusUpdates[1])
	}
}

// TestHandleReceiptEventIgnoresUnrelated memverifikasi receipt tipe lain
// (sender/retry/played) tidak mengubah status centang pesan.
func TestHandleReceiptEventIgnoresUnrelated(t *testing.T) {
	repo := newFakeHistoryRepo()
	c := &Client{SessionID: 7, chatRepo: repo, waClient: &whatsmeow.Client{}}

	c.handleReceiptEvent(&events.Receipt{
		MessageSource: types.MessageSource{
			Chat: types.NewJID("628123456789", types.DefaultUserServer),
		},
		MessageIDs: []types.MessageID{"id-played"},
		Type:       types.ReceiptTypePlayed,
	})

	repo.mu.Lock()
	defer repo.mu.Unlock()
	if len(repo.statusUpdates) != 0 {
		t.Errorf("statusUpdates = %d, want 0 (played tidak mengubah centang)", len(repo.statusUpdates))
	}
}
