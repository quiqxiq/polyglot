package whatsapp

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/skip2/go-qrcode"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
	"google.golang.org/protobuf/proto"

	"github.com/quixiq/polyglot/internal/domain/bot"
	"github.com/quixiq/polyglot/internal/port"
)

// MessageCallback meneruskan pesan masuk ke handler (Engine bot). chatJID
// adalah JID chat (1:1 = nomor@s.whatsapp.net, grup = @g.us) — dipakai untuk
// kontrol bot per-chat; customerNumber adalah nomor pengirim (tanpa @server).
type MessageCallback func(sessionID uint, chatJID string, customerNumber string, content string)
type StatusCallback func(sessionID uint, status string, qrCode string, jid string, phoneNumber string)

// ChatUpdateCallback memberitahukan bahwa mirror chat berubah (pesan masuk
// ATAU keluar sudah ditulis ke wa_messages/wa_chats) — dipakai untuk broadcast
// SSE `chat_update` agar Inbox frontend ter-update instan tanpa polling.
type ChatUpdateCallback func(sessionID uint, chatJID string)

func init() {
	store.SetOSInfo("Polyglot NetOps WA Bot", [3]uint32{2, 3000, 1015901307})
}

type Client struct {
	SessionID   uint
	deviceStore *store.Device
	waClient    *whatsmeow.Client
	qrCode      string
	qrBase64    string
	qrMutex     sync.RWMutex
	onMessage   MessageCallback
	callbackMu  sync.RWMutex
	onStatus    StatusCallback
	onChatUpd   ChatUpdateCallback
	chatRepo    port.ChatRepository

	// qrFlowMu/qrFlowActive memastikan hanya ada SATU aliran QR (GetQRChannel +
	// goroutine watcher) per client. Tanpa guard ini, polling GetQRCode yang
	// sering (tiap 2-5s) bisa menumpuk channel & goroutine.
	qrFlowMu     sync.Mutex
	qrFlowActive bool
}

func NewClient(sessionID uint, deviceStore *store.Device, onMsg MessageCallback, onStat StatusCallback, chatRepo port.ChatRepository) *Client {
	logger := waLog.Stdout("WhatsApp", "INFO", true)
	client := whatsmeow.NewClient(deviceStore, logger)
	client.EnableAutoReconnect = true
	client.AutoTrustIdentity = true

	c := &Client{
		SessionID:   sessionID,
		deviceStore: deviceStore,
		waClient:    client,
		onMessage:   onMsg,
		onStatus:    onStat,
		chatRepo:    chatRepo,
	}

	client.AddEventHandler(c.handleEvent)
	return c
}

// SetMessageCallback swaps the incoming-message handler at runtime. Digunakan
// setelah Engine bot selesai di-bootstrap (app wiring) untuk menghindari
// circular dependency SessionManager <-> Engine.
func (c *Client) SetMessageCallback(cb MessageCallback) {
	c.callbackMu.Lock()
	defer c.callbackMu.Unlock()
	c.onMessage = cb
}

func (c *Client) getMessageCallback() MessageCallback {
	c.callbackMu.RLock()
	defer c.callbackMu.RUnlock()
	return c.onMessage
}

// SetChatUpdateCallback swaps the chat-mirror-change handler at runtime.
// Dipasang lewat SessionManager.SetChatUpdateCallback setelah EventHandler
// (pemilik SSE hub) dibangun.
func (c *Client) SetChatUpdateCallback(cb ChatUpdateCallback) {
	c.callbackMu.Lock()
	defer c.callbackMu.Unlock()
	c.onChatUpd = cb
}

func (c *Client) getChatUpdateCallback() ChatUpdateCallback {
	c.callbackMu.RLock()
	defer c.callbackMu.RUnlock()
	return c.onChatUpd
}

func (c *Client) notifyChatUpdate(chatJID string) {
	if cb := c.getChatUpdateCallback(); cb != nil {
		cb(c.SessionID, chatJID)
	}
}

func (c *Client) Connect(ctx context.Context) error {
	if c.waClient.IsConnected() {
		// Websocket sudah hidup. Bila belum ter-pair namun aliran QR mati
		// (mis. QR timeout), hidupkan ulang supaya GetQRCode berikutnya
		// mendapat QR baru tanpa perlu tombol Reconnect manual.
		if c.waClient.Store.ID == nil {
			c.EnsureQRFlow()
		}
		return nil
	}

	if c.waClient.Store.ID == nil {
		// Session belum pernah ter-pair → jalankan QR flow.
		c.startQRFlow()
		return nil
	}

	if err := c.waClient.Connect(); err != nil {
		return fmt.Errorf("failed to connect WA client: %w", err)
	}
	if c.onStatus != nil {
		userJID := ""
		phoneNum := ""
		if c.waClient.Store.ID != nil {
			userJID = c.waClient.Store.ID.String()
			phoneNum = c.waClient.Store.ID.User
		}
		c.onStatus(c.SessionID, "online", "", userJID, phoneNum)
	}

	return nil
}

// EnsureQRFlow memastikan aliran QR aktif untuk session yang belum ter-pair.
// Idempoten — dipanggil dari Connect dan GetQRCode (polling 2-5s) tanpa risiko
// menumpuk GetQRChannel/goroutine.
func (c *Client) EnsureQRFlow() {
	if c.waClient.Store.ID != nil {
		return
	}
	c.startQRFlow()
}

// startQRFlow membuka channel QR (detached 3-menit context agar HTTP timeout
// tidak membatalkan polling QR) dan menyalakan goroutine watcher. Guard
// qrFlowActive menjamin hanya satu flow berjalan; saat channel ditutup
// (timeout/error), flag di-reset sehingga poll berikutnya memicu flow baru.
func (c *Client) startQRFlow() {
	c.qrFlowMu.Lock()
	if c.qrFlowActive {
		c.qrFlowMu.Unlock()
		return
	}
	c.qrFlowActive = true
	c.qrFlowMu.Unlock()

	if c.onStatus != nil {
		c.onStatus(c.SessionID, "connecting", "", "", "")
	}
	c.waClient.Disconnect()

	qrCtx, qrCancel := context.WithTimeout(context.Background(), 3*time.Minute)
	qrChan, err := c.waClient.GetQRChannel(qrCtx)
	if err != nil {
		qrCancel()
		c.qrFlowMu.Lock()
		c.qrFlowActive = false
		c.qrFlowMu.Unlock()
		if errors.Is(err, whatsmeow.ErrQRStoreContainsID) {
			log.Printf("[WhatsApp Client %d] Store already contains ID, attempting direct connect...", c.SessionID)
			if err := c.waClient.Connect(); err != nil {
				log.Printf("[WhatsApp Client %d] connect with saved session failed: %v", c.SessionID, err)
			}
			return
		}
		log.Printf("[WhatsApp Client %d] failed to get QR channel: %v", c.SessionID, err)
		if c.onStatus != nil {
			c.onStatus(c.SessionID, "needs_rescan", "", "", "")
		}
		return
	}

	if !c.waClient.IsConnected() {
		if err := c.waClient.Connect(); err != nil {
			qrCancel()
			c.qrFlowMu.Lock()
			c.qrFlowActive = false
			c.qrFlowMu.Unlock()
			if c.onStatus != nil {
				c.onStatus(c.SessionID, "needs_rescan", "", "", "")
			}
			log.Printf("[WhatsApp Client %d] failed to connect WA client: %v", c.SessionID, err)
			return
		}
	}

	go func() {
		defer qrCancel()
		defer func() {
			// Channel QR ditutup (timeout / flow selesai) → buka slot untuk
			// restart QR berikutnya.
			c.qrFlowMu.Lock()
			c.qrFlowActive = false
			c.qrFlowMu.Unlock()
		}()
		for evt := range qrChan {
			if evt.Event == "code" {
				c.qrMutex.Lock()
				c.qrCode = evt.Code
				if pngBytes, err := qrcode.Encode(evt.Code, qrcode.Medium, 256); err == nil {
					c.qrBase64 = "data:image/png;base64," + base64.StdEncoding.EncodeToString(pngBytes)
				}
				c.qrMutex.Unlock()

				if c.onStatus != nil {
					c.onStatus(c.SessionID, "needs_rescan", c.qrBase64, "", "")
				}
			} else if evt.Event == "timeout" {
				// QR kedaluwarsa: QR cache dikosongkan dan status needs_rescan
				// di-broadcast. Channel akan ditutup oleh whatsmeow → defer
				// me-reset qrFlowActive → GetQRCode berikutnya me-restart flow.
				log.Printf("[WhatsApp Client %d] QR code expired", c.SessionID)
				c.qrMutex.Lock()
				c.qrCode = ""
				c.qrBase64 = ""
				c.qrMutex.Unlock()
				if c.onStatus != nil {
					c.onStatus(c.SessionID, "needs_rescan", "", "", "")
				}
			} else {
				log.Printf("[WhatsApp Client %d] QR event: %s", c.SessionID, evt.Event)
			}
		}
	}()
}

func (c *Client) Disconnect() {
	if c.waClient.IsConnected() {
		c.waClient.Disconnect()
	}
}

func (c *Client) Logout(ctx context.Context) error {
	// Best-effort unlink dari server WhatsApp. Kegagalan tidak memblokir
	// pembersihan lokal (keep-slot di DB).
	if c.waClient.IsConnected() {
		if err := c.waClient.Logout(ctx); err != nil {
			log.Printf("[WhatsApp Client %d] Remote logout request failed (best-effort): %v", c.SessionID, err)
		}
	}
	c.Disconnect()
	// Hapus session lokal supaya slot bisa di-pair ulang dengan QR baru.
	if c.deviceStore != nil {
		_ = c.deviceStore.Delete(ctx)
	}
	return nil
}

func (c *Client) Reconnect(ctx context.Context) error {
	log.Printf("[WhatsApp Client %d] Manual reconnect initiated...", c.SessionID)
	c.Disconnect()
	time.Sleep(1 * time.Second)
	return c.Connect(ctx)
}

func (c *Client) GetQRCode() string {
	c.qrMutex.RLock()
	defer c.qrMutex.RUnlock()
	if c.qrBase64 != "" {
		return c.qrBase64
	}
	return c.qrCode
}

func (c *Client) GetPairingCode(ctx context.Context, phoneNumber string) (string, error) {
	if !c.waClient.IsConnected() {
		if err := c.waClient.Connect(); err != nil {
			return "", fmt.Errorf("failed to connect for pairing: %w", err)
		}
	}
	code, err := c.waClient.PairPhone(ctx, phoneNumber, true, whatsmeow.PairClientChrome, "Chrome (Polyglot NetOps)")
	if err != nil {
		return "", fmt.Errorf("failed to request pairing code: %w", err)
	}
	return code, nil
}

func (c *Client) SendMessage(ctx context.Context, to string, content string) error {
	if !c.waClient.IsConnected() {
		return fmt.Errorf("WA client for session %d is disconnected", c.SessionID)
	}

	jid, err := parseJID(to)
	if err != nil {
		return fmt.Errorf("invalid recipient JID (%s): %w", to, err)
	}

	msg := &waE2E.Message{
		Conversation: proto.String(content),
	}

	respID, err := c.waClient.SendMessage(ctx, jid, msg)
	if err != nil {
		return fmt.Errorf("failed to send WA message: %w", err)
	}

	c.recordOutgoingMessage(jid, string(respID.ID), content, "text")

	return nil
}

func (c *Client) SendDocument(ctx context.Context, to string, fileBytes []byte, fileName string, contentType string, caption string) error {
	if !c.waClient.IsConnected() {
		return fmt.Errorf("WA client for session %d is disconnected", c.SessionID)
	}

	jid, err := parseJID(to)
	if err != nil {
		return fmt.Errorf("invalid recipient JID (%s): %w", to, err)
	}

	if contentType == "" {
		contentType = http.DetectContentType(fileBytes)
		if contentType == "text/plain; charset=utf-8" && strings.HasSuffix(strings.ToLower(fileName), ".pdf") {
			contentType = "application/pdf"
		}
	}

	uploadResp, err := c.waClient.Upload(ctx, fileBytes, whatsmeow.MediaDocument)
	if err != nil {
		return fmt.Errorf("failed to upload document to WA servers: %w", err)
	}

	docMsg := &waE2E.DocumentMessage{
		URL:           proto.String(uploadResp.URL),
		DirectPath:    proto.String(uploadResp.DirectPath),
		MediaKey:      uploadResp.MediaKey,
		Mimetype:      proto.String(contentType),
		FileSHA256:    uploadResp.FileSHA256,
		FileEncSHA256: uploadResp.FileEncSHA256,
		FileLength:    proto.Uint64(uploadResp.FileLength),
		FileName:      proto.String(fileName),
	}
	if caption != "" {
		docMsg.Caption = proto.String(caption)
	}

	msg := &waE2E.Message{
		DocumentMessage: docMsg,
	}

	respID, err := c.waClient.SendMessage(ctx, jid, msg)
	if err != nil {
		return fmt.Errorf("failed to send WA document message: %w", err)
	}

	c.recordOutgoingMessage(jid, string(respID.ID), caption, "document")

	return nil
}

func (c *Client) SendImage(ctx context.Context, to string, imageBytes []byte, contentType string, caption string) error {
	if !c.waClient.IsConnected() {
		return fmt.Errorf("WA client for session %d is disconnected", c.SessionID)
	}

	jid, err := parseJID(to)
	if err != nil {
		return fmt.Errorf("invalid recipient JID (%s): %w", to, err)
	}

	if contentType == "" {
		contentType = http.DetectContentType(imageBytes)
	}

	uploadResp, err := c.waClient.Upload(ctx, imageBytes, whatsmeow.MediaImage)
	if err != nil {
		return fmt.Errorf("failed to upload image to WA servers: %w", err)
	}

	imgMsg := &waE2E.ImageMessage{
		URL:           proto.String(uploadResp.URL),
		DirectPath:    proto.String(uploadResp.DirectPath),
		MediaKey:      uploadResp.MediaKey,
		Mimetype:      proto.String(contentType),
		FileSHA256:    uploadResp.FileSHA256,
		FileEncSHA256: uploadResp.FileEncSHA256,
		FileLength:    proto.Uint64(uploadResp.FileLength),
	}
	if caption != "" {
		imgMsg.Caption = proto.String(caption)
	}

	msg := &waE2E.Message{
		ImageMessage: imgMsg,
	}

	respID, err := c.waClient.SendMessage(ctx, jid, msg)
	if err != nil {
		return fmt.Errorf("failed to send WA image message: %w", err)
	}

	c.recordOutgoingMessage(jid, string(respID.ID), caption, "image")

	return nil
}

func (c *Client) handleEvent(evt any) {
	switch v := evt.(type) {
	case *events.Message:
		c.persistMirrorMessage(v)
		if !v.Info.IsFromMe {
			c.handleIncomingMessage(v)
		}
	case *events.Connected:
		log.Printf("[WhatsApp Client %d] Connected successfully (Auto-reconnect active)", c.SessionID)
		// Saat sudah ter-pair, QR tidak relevan lagi — kosongkan cache agar UI
		// tidak menampilkan QR basi untuk device yang sudah online.
		c.qrMutex.Lock()
		c.qrCode = ""
		c.qrBase64 = ""
		c.qrMutex.Unlock()
		if c.onStatus != nil && c.waClient.Store.ID != nil {
			c.onStatus(c.SessionID, "online", "", c.waClient.Store.ID.String(), c.waClient.Store.ID.User)
		}
	case *events.Disconnected:
		log.Printf("[WhatsApp Client %d] Disconnected (Auto-reconnect will retry)", c.SessionID)
		if c.onStatus != nil {
			c.onStatus(c.SessionID, "offline", "", "", "")
		}
	case *events.LoggedOut:
		log.Printf("[WhatsApp Client %d] Remote LoggedOut event received from phone", c.SessionID)
		// Remote logout (device di-unlink dari HP): hapus session lokal supaya
		// slot bisa di-pair ulang. Slot DB dipertahankan (keep-slot), sama seperti
		// referensi go-whatsapp-web-multidevice.
		if c.deviceStore != nil {
			_ = c.deviceStore.Delete(context.Background())
		}
		if c.onStatus != nil {
			c.onStatus(c.SessionID, "needs_rescan", "", "", "")
		}
	}
}

func (c *Client) handleIncomingMessage(evt *events.Message) {
	if evt.Info.IsFromMe {
		return
	}

	senderJID := evt.Info.Sender.User
	if senderJID == "" {
		senderJID = evt.Info.Chat.User
	}

	chatJID := evt.Info.Chat.String()
	if chatJID == "" {
		return
	}

	body := extractMessageBody(evt.Message)
	if body == "" {
		return
	}

	log.Printf("[WhatsApp Client %d] Message from %s (chat %s): %s", c.SessionID, senderJID, chatJID, body)

	if cb := c.getMessageCallback(); cb != nil {
		cb(c.SessionID, chatJID, senderJID, body)
	}
}

func extractMessageBody(msg *waE2E.Message) string {
	if msg == nil {
		return ""
	}
	if msg.GetConversation() != "" {
		return msg.GetConversation()
	}
	if msg.ExtendedTextMessage != nil && msg.ExtendedTextMessage.GetText() != "" {
		return msg.ExtendedTextMessage.GetText()
	}
	if msg.ImageMessage != nil && msg.ImageMessage.GetCaption() != "" {
		return msg.ImageMessage.GetCaption()
	}
	if msg.VideoMessage != nil && msg.VideoMessage.GetCaption() != "" {
		return msg.VideoMessage.GetCaption()
	}
	return ""
}

// extractMediaType maps a waE2E message to a coarse media type label used by
// the Inbox UI (text/image/video/audio/document/...).
func extractMediaType(msg *waE2E.Message) string {
	if msg == nil {
		return "text"
	}
	switch {
	case msg.GetConversation() != "" || msg.ExtendedTextMessage != nil:
		return "text"
	case msg.ImageMessage != nil:
		return "image"
	case msg.VideoMessage != nil:
		return "video"
	case msg.AudioMessage != nil:
		return "audio"
	case msg.DocumentMessage != nil || msg.DocumentWithCaptionMessage != nil:
		return "document"
	case msg.StickerMessage != nil:
		return "sticker"
	case msg.LocationMessage != nil || msg.LiveLocationMessage != nil:
		return "location"
	case msg.ContactMessage != nil || msg.ContactsArrayMessage != nil:
		return "contact"
	default:
		return "unknown"
	}
}

// previewOf builds the chat-list preview text for a message.
func previewOf(content, mediaType string) string {
	if content != "" {
		return content
	}
	switch mediaType {
	case "image":
		return "📷 Foto"
	case "video":
		return "🎬 Video"
	case "audio":
		return "🎵 Audio"
	case "document":
		return "📄 Dokumen"
	case "sticker":
		return "🖼️ Stiker"
	case "location":
		return "📍 Lokasi"
	case "contact":
		return "👤 Kontak"
	case "call":
		return "📞 Panggilan"
	default:
		return "[media]"
	}
}

// persistMirrorMessage writes every WhatsApp message (incoming AND outgoing)
// into the Inbox mirror tables so the UI can show the full conversation list.
func (c *Client) persistMirrorMessage(evt *events.Message) {
	if c.chatRepo == nil {
		return
	}

	chatJID := evt.Info.Chat.String()
	if chatJID == "" {
		return
	}

	// Untuk 1:1, display name memakai push name pengirim. Untuk grup, PushName
	// adalah nama pengirim (bukan nama grup) — limitation Fase A: nama grup
	// akurat bisa didapat dari contact/group store whatsmeow bila dibutuhkan.
	displayName := evt.Info.PushName
	if !evt.Info.IsGroup && displayName == "" && evt.Info.Sender.User != "" {
		displayName = evt.Info.Sender.User
	}

	content := extractMessageBody(evt.Message)
	mediaType := extractMediaType(evt.Message)

	msg := &bot.WAMessage{
		SessionID:   c.SessionID,
		ChatJID:     chatJID,
		WAMessageID: evt.Info.ID,
		SenderJID:   evt.Info.Sender.String(),
		SenderName:  evt.Info.PushName,
		Content:     content,
		MediaType:   mediaType,
		IsFromMe:    evt.Info.IsFromMe,
		Timestamp:   evt.Info.Timestamp,
	}
	if msg.WAMessageID == "" {
		msg.WAMessageID = fmt.Sprintf("evt-%d", evt.Info.Timestamp.UnixNano())
	}

	// Tiga write terpisah (pesan → chat → unread) tanpa transaksi: event per
	// client diproses serial oleh whatsmeow, jadi race antar-tulis tidak terjadi;
	// kegagalan tengah-jalan hanya menyisakan preview chat yang tertinggal satu
	// pesan dan akan terkoreksi oleh event berikutnya.
	inserted, err := c.chatRepo.UpsertMessage(msg)
	if err != nil {
		log.Printf("[WhatsApp Client %d] Failed to mirror message: %v", c.SessionID, err)
	}

	chat := &bot.WAChat{
		SessionID:          c.SessionID,
		ChatJID:            chatJID,
		DisplayName:        displayName,
		IsGroup:            evt.Info.IsGroup,
		LastMessageID:      msg.WAMessageID,
		LastMessagePreview: previewOf(content, mediaType),
		LastMessageTime:    evt.Info.Timestamp,
	}
	if err := c.chatRepo.UpsertChat(chat); err != nil {
		log.Printf("[WhatsApp Client %d] Failed to mirror chat: %v", c.SessionID, err)
	}

	// Unread hanya dihitung saat pesan BARU benar-benar masuk (insert, bukan
	// event ulang) — mencegah dobel-hitungan unread pada replay event.
	if inserted && !evt.Info.IsFromMe {
		if err := c.chatRepo.IncrementUnread(c.SessionID, chatJID); err != nil {
			log.Printf("[WhatsApp Client %d] Failed to increment unread: %v", c.SessionID, err)
		}
	}

	// Beri tahu UI (via SSE) bahwa mirror chat berubah — Inbox refresh instan.
	c.notifyChatUpdate(chatJID)
}

// recordOutgoingMessage mirrors messages sent from this device (bot/agent) into
// the Inbox so sent messages appear in the conversation view immediately.
func (c *Client) recordOutgoingMessage(jid types.JID, waMessageID string, content string, mediaType string) {
	if c.chatRepo == nil {
		return
	}

	chatJID := jid.String()
	if chatJID == "" {
		return
	}
	if waMessageID == "" {
		waMessageID = fmt.Sprintf("out-%d", time.Now().UnixNano())
	}
	now := time.Now()

	msg := &bot.WAMessage{
		SessionID:   c.SessionID,
		ChatJID:     chatJID,
		WAMessageID: waMessageID,
		Content:     content,
		MediaType:   mediaType,
		IsFromMe:    true,
		IsRead:      true,
		Timestamp:   now,
	}
	if _, err := c.chatRepo.UpsertMessage(msg); err != nil {
		log.Printf("[WhatsApp Client %d] Failed to mirror outgoing message: %v", c.SessionID, err)
	}

	chat := &bot.WAChat{
		SessionID:          c.SessionID,
		ChatJID:            chatJID,
		LastMessageID:      waMessageID,
		LastMessagePreview: previewOf(content, mediaType),
		LastMessageTime:    now,
	}
	if err := c.chatRepo.UpsertChat(chat); err != nil {
		log.Printf("[WhatsApp Client %d] Failed to mirror outgoing chat: %v", c.SessionID, err)
	}

	// Beri tahu UI (via SSE) bahwa mirror chat berubah — balasan bot/agen
	// langsung tampil di Inbox tanpa polling.
	c.notifyChatUpdate(chatJID)
}

func parseJID(target string) (types.JID, error) {
	target = strings.TrimSpace(target)
	if strings.Contains(target, "@") {
		return types.ParseJID(target)
	}

	cleanNum := strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, target)

	if strings.HasPrefix(cleanNum, "0") {
		cleanNum = "62" + cleanNum[1:]
	}

	return types.NewJID(cleanNum, types.DefaultUserServer), nil
}
