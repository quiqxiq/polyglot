package whatsapp

import (
	"context"
	"github.com/quixiq/polyglot/pkg/logger"
	"time"

	"go.mau.fi/whatsmeow/proto/waHistorySync"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"

	"github.com/quixiq/polyglot/internal/domain/bot"
)

// historyMessageBatchSize membatasi jumlah baris per statement INSERT multi-row
// saat mem-mirror history sync — cukup besar untuk memangkas round-trip, cukup
// kecil untuk menjaga ukuran statement tetap wajar.
const historyMessageBatchSize = 500

// handleHistorySync mirrors WhatsApp history sync blobs into the Inbox tables
// (wa_chats/wa_messages) sehingga SEMUA chat — termasuk grup — langsung muncul
// di halaman Chats setelah device ter-pair, tanpa menunggu pesan live baru.
// Dipanggil dari Client.handleEvent (events.HistorySync) di dalam goroutine
// agar event loop whatsmeow tidak terblokir oleh ribuan insert history sync.
func (c *Client) handleHistorySync(evt *events.HistorySync) {
	if evt == nil || evt.Data == nil || c.chatRepo == nil {
		return
	}
	go func() {
		// Data history sync berasal dari server WhatsApp (untrusted) — panic di
		// goroutine latar belakang akan meruntuhkan seluruh proses, jadi di-recover
		// dan dicatat saja.
		defer func() {
			if r := recover(); r != nil {
				logger.WithComponent("HistorySync").Errorf("Panic during history sync mirror (session %d): %v", c.SessionID, r)
			}
		}()
		switch evt.Data.GetSyncType() {
		case waHistorySync.HistorySync_INITIAL_BOOTSTRAP, waHistorySync.HistorySync_RECENT:
			c.processHistoryConversations(evt.Data)
		case waHistorySync.HistorySync_FULL, waHistorySync.HistorySync_ON_DEMAND:
			// FULL terjadi saat user membuka chat di HP (lazy load); ON_DEMAND
			// saat aplikasi merequest sync manual — keduanya membawa conversations
			// nyata yang harus di-mirror. Referensi memproses keduanya.
			c.processHistoryConversations(evt.Data)
		case waHistorySync.HistorySync_PUSH_NAME:
			c.processHistoryPushNames(evt.Data)
		default:
			// Tipe lain yang genuinely tidak membawa conversations.
		}
	}()
}

// processHistoryConversations mirrors conversations (chat + pesannya) dari
// history sync. Upsert idempotent per (session_id, wa_message_id), sehingga
// aman dipanggil berulang (INITIAL_BOOTSTRAP lalu RECENT) tanpa duplikasi.
func (c *Client) processHistoryConversations(data *waHistorySync.HistorySync) {
	conversations := data.GetConversations()
	if len(conversations) == 0 {
		return
	}
	logger.WithComponent("HistorySync").Infof("Session %d: mirroring %d conversations", c.SessionID, len(conversations))

	selfJID := ""
	if c.waClient != nil && c.waClient.Store != nil && c.waClient.Store.ID != nil {
		selfJID = c.waClient.Store.ID.ToNonAD().String()
	}

	for _, conv := range conversations {
		rawJID := conv.GetID()
		if rawJID == "" {
			continue
		}
		jid, err := types.ParseJID(rawJID)
		if err != nil {
			logger.WithComponent("HistorySync").Warnf("Session %d: skip chat %q (invalid JID): %v", c.SessionID, rawJID, err)
			continue
		}
		// Fase 1: normalisasi LID → nomor HP untuk CHAT JID (bukan hanya sender).
		// WhatsApp mengirim history sync dengan @lid untuk kontak privasi;
		// tanpa ini chat tampil sebagai angka LID di Inbox.
		jid = normalizeJIDFromLID(context.Background(), jid, c.waClient)
		chatJID := jid.String()

		// Baris chat @lid basi (dari sync sebelum LID map tersedia) digabung
		// ke baris nomor HP-nya — mencegah chat dobel yang tidak cocok dengan HP.
		c.reconcileLIDChat(context.Background(), jid)

		// Fase 1: sistem JID (story & channel) tidak masuk mirror Inbox.
		if isSkippedJID(chatJID) {
			continue
		}

		isGroup := jid.Server == types.GroupServer

		var latest time.Time
		var latestPreview string
		var lastMsgID string
		fallbackName := ""

		// Pesan dikumpulkan lalu di-flush ber-batch (historyMessageBatchSize)
		// — satu statement INSERT multi-row per chunk, bukan satu query per
		// pesan. Tanpa ini INITIAL_BOOTSTRAP dengan ribuan pesan = ribuan
		// round-trip ke database.
		var batch []*bot.WAMessage
		flush := func() {
			if len(batch) == 0 {
				return
			}
			if _, err := c.chatRepo.UpsertMessagesBatch(context.Background(), batch); err != nil {
				logger.WithComponent("HistorySync").Errorf("Session %d: failed to mirror %d messages of %s: %v", c.SessionID, len(batch), chatJID, err)
			}
			batch = batch[:0]
		}

		for _, histMsg := range conv.GetMessages() {
			if histMsg == nil || histMsg.Message == nil {
				continue
			}
			info := histMsg.Message
			key := info.GetKey()
			if key == nil || key.GetID() == "" {
				continue
			}
			// Pesan tidak terdekripsi / placeholder tanpa isi — tidak ada yang
			// bisa di-mirror (extractMediaType(nil) mengembalikan "text").
			if info.GetMessage() == nil {
				continue
			}

			body := extractMessageBody(info.GetMessage())
			mediaType := extractMediaType(info.GetMessage())
			// Pesan sistem (stub) / reaction / protocol tanpa konten & tanpa media
			// tak perlu di-mirror — hanya memenuhi daftar chat dengan noise
			// (termasuk "[media]" yang tidak ada di HP).
			if body == "" && (mediaType == "unknown" || mediaType == "reaction" || mediaType == "system") {
				continue
			}

			isFromMe := key.GetFromMe()
			senderJIDStr := ""
			switch {
			case isFromMe:
				senderJIDStr = selfJID
			case key.GetParticipant() != "":
				// Fase 4: normalisasi LID → nomor HP untuk sender.
				parsedSender, perr := types.ParseJID(key.GetParticipant())
				if perr == nil {
					senderJIDStr = normalizeJIDFromLID(context.Background(), parsedSender, c.waClient).String()
				} else {
					senderJIDStr = key.GetParticipant()
				}
			case info.GetParticipant() != "":
				// History-sync group messages menaruh pengirim di level
				// WebMessageInfo, bukan key.participant (lihat referensi
				// go-whatsapp-web-multidevice, issue #609).
				parsedSender, perr := types.ParseJID(info.GetParticipant())
				if perr == nil {
					senderJIDStr = normalizeJIDFromLID(context.Background(), parsedSender, c.waClient).String()
				} else {
					senderJIDStr = info.GetParticipant()
				}
			case isGroup:
				// Pesan grup tanpa info pengirim tidak bisa diatribusikan.
				continue
			default:
				senderJIDStr = chatJID
			}

			ts := time.Unix(int64(info.GetMessageTimestamp()), 0)
			batch = append(batch, &bot.WAMessage{
				SessionID:   c.SessionID,
				ChatJID:     chatJID,
				WAMessageID: key.GetID(),
				SenderJID:   senderJIDStr,
				SenderName:  info.GetPushName(),
				Content:     body,
				MediaType:   mediaType,
				IsFromMe:    isFromMe,
				Timestamp:   ts,
			})
			if len(batch) >= historyMessageBatchSize {
				flush()
			}
			if ts.After(latest) {
				latest = ts
				latestPreview = previewOf(body, mediaType)
				lastMsgID = key.GetID()
			}
			if fallbackName == "" && info.GetPushName() != "" {
				fallbackName = info.GetPushName()
			}
		}
		flush()

		displayName := c.resolveChatDisplayName(context.Background(), jid, conv.GetName())
		if displayName == "" {
			displayName = fallbackName
		}
		if displayName == "" {
			displayName = jid.User
		}

		chat := &bot.WAChat{
			SessionID:          c.SessionID,
			ChatJID:            chatJID,
			DisplayName:        displayName,
			IsGroup:            isGroup,
			LastMessageID:      lastMsgID,
			LastMessagePreview: latestPreview,
			LastMessageTime:    latest,
		}
		if err := c.chatRepo.UpsertChat(context.Background(), chat); err != nil {
			logger.WithComponent("HistorySync").Errorf("Session %d: failed to mirror chat %s: %v", c.SessionID, chatJID, err)
		} else {
			// History sync otoritatif untuk unread — set selalu (termasuk 0)
			// agar angka di DB tidak menyimpan nilai basi dari sinkronisasi
			// sebelumnya.
			if err := c.chatRepo.SetChatUnread(context.Background(), c.SessionID, chatJID, conv.GetUnreadCount()); err != nil {
				logger.WithComponent("HistorySync").Warnf("Session %d: failed to set unread for %s: %v", c.SessionID, chatJID, err)
			}
		}
		// Beri tahu UI (via SSE) bahwa mirror chat berubah — Inbox refresh.
		c.notifyChatUpdate(chatJID)
	}
}

// processHistoryPushNames mirrors push names (nama yang diset pengguna kontak)
// ke chat 1:1. Untuk chat yang sudah ada, display_name di-update; untuk kontak
// yang belum punya chat, dibuat baris chat ringan dengan LastMessageTime kosong
// sehingga kontak tetap tampil di daftar (tersortir paling bawah).
//
// Keterbatasan yang diketahui: UpsertChat selalu menimpa display_name, jadi
// nama kontak kustom yang disimpan pengguna bisa tergantikan push name —
// perilaku yang sama dengan referensi go-whatsapp-web-multidevice.
func (c *Client) processHistoryPushNames(data *waHistorySync.HistorySync) {
	pushnames := data.GetPushnames()
	if len(pushnames) == 0 {
		return
	}
	for _, pn := range pushnames {
		if pn == nil {
			continue
		}
		rawJID := pn.GetID()
		name := pn.GetPushname()
		if rawJID == "" || name == "" {
			continue
		}
		jid, err := types.ParseJID(rawJID)
		if err != nil {
			continue
		}
		// Fase 1: normalisasi LID → nomor HP juga untuk push name, supaya
		// baris chat yang di-update match dengan chat 1:1 (yang sudah
		// dinormalisasi) — bukan baris @lid yang tidak pernah ada.
		jid = normalizeJIDFromLID(context.Background(), jid, c.waClient)
		// Baris chat @lid basi untuk kontak ini juga digabung ke nomor HP.
		c.reconcileLIDChat(context.Background(), jid)
		chat := &bot.WAChat{
			SessionID:   c.SessionID,
			ChatJID:     jid.String(),
			DisplayName: name,
		}
		if err := c.chatRepo.UpsertChat(context.Background(), chat); err != nil {
			logger.WithComponent("HistorySync").Warnf("Session %d: failed to save push name for %s: %v", c.SessionID, rawJID, err)
		}
	}
}
