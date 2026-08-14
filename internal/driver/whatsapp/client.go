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

// ChatPresenceCallback memberitahukan event typing/recording dari kontak.
// state: "composing" atau "paused"; media: "" (teks) atau "audio" (voice).
type ChatPresenceCallback func(sessionID uint, chatJID, senderJID, state, media string, isGroup bool)

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
	onChatPres  ChatPresenceCallback
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

// SetChatPresenceCallback swaps the chat-presence handler at runtime.
// Dipasang lewat SessionManager.SetChatPresenceCallback setelah EventHandler
// dibangun.
func (c *Client) SetChatPresenceCallback(cb ChatPresenceCallback) {
	c.callbackMu.Lock()
	defer c.callbackMu.Unlock()
	c.onChatPres = cb
}

func (c *Client) getChatPresenceCallback() ChatPresenceCallback {
	c.callbackMu.RLock()
	defer c.callbackMu.RUnlock()
	return c.onChatPres
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

// sendPresenceAvailable menandai device online ke server WhatsApp. Dipanggil
// saat Connected dan AppStateSyncComplete — kehadiran presence membuat HP
// mengalirkan history sync (chat lama) ke perangkat tertaut, sehingga Inbox
// tidak hanya berisi beberapa chat terbaru.
func (c *Client) sendPresenceAvailable() {
	if c.waClient == nil || !c.waClient.IsConnected() || c.waClient.Store == nil || c.waClient.Store.ID == nil {
		return
	}
	if err := c.waClient.SendPresence(context.Background(), types.PresenceAvailable); err != nil {
		log.Printf("[WhatsApp Client %d] failed to send presence: %v", c.SessionID, err)
		return
	}
	log.Printf("[WhatsApp Client %d] presence available sent (history sync trigger)", c.SessionID)
}

// reconcileLIDChat mencegah chat @lid basi mendobel chat nomor HP-nya.
// WhatsApp bisa mengirim history sync berformat @lid sebelum LID map tersedia
// (app state); baris @lid yang tersimpan lalu tidak pernah terpakai lagi dan
// tampil sebagai nomor LID yang tidak dikenal di Inbox — tidak sama dengan HP.
// Dipanggil setiap kali sebuah chat 1:1 di-normalisasi ke nomor HP.
func (c *Client) reconcileLIDChat(ctx context.Context, resolved types.JID) {
	if c.chatRepo == nil || c.waClient == nil || c.waClient.Store == nil || c.waClient.Store.LIDs == nil {
		return
	}
	lid, err := c.waClient.Store.LIDs.GetLIDForPN(ctx, resolved)
	if err != nil || lid.IsEmpty() || lid.Server != "lid" {
		return
	}
	if err := c.chatRepo.MergeChatLID(c.SessionID, lid.String(), resolved.String()); err != nil {
		log.Printf("[WhatsApp Client %d] failed to merge stale LID chat %s → %s: %v", c.SessionID, lid.String(), resolved.String(), err)
	}
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
	case *events.Receipt:
		// Acknowledge pesan keluar: "delivered" (✓✓) saat sampai di device
		// penerima, "read" (✓✓ biru) saat dibaca — 4 status WhatsApp.
		c.handleReceiptEvent(v)
	case *events.ChatPresence:
		c.handleChatPresenceEvent(v)
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
		// Tandai device online — tanpa presence, HP menganggap perangkat idle
		// dan hanya mengirim history sync terbatas.
		c.sendPresenceAvailable()
	case *events.AppStateSyncComplete:
		// App state selesai di-resync (tepat setelah connect/pair) — kirim
		// presence available agar HP mulai mengalirkan history sync lengkap.
		// whatsmeow tidak mengirim presence otomatis; pola yang sama dipakai
		// referensi go-whatsapp-web-multidevice (presence pulse berkala).
		c.sendPresenceAvailable()
	case *events.Disconnected:
		log.Printf("[WhatsApp Client %d] Disconnected (Auto-reconnect will retry)", c.SessionID)
		if c.onStatus != nil {
			c.onStatus(c.SessionID, "offline", "", "", "")
		}
	case *events.HistorySync:
		// Mirror semua chat + pesan (termasuk grup) dari HP ke Inbox agar
		// halaman Chats menampilkan seluruh percakapan, bukan cuma yang
		// menerima pesan live setelah terhubung.
		c.handleHistorySync(v)
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

	chatJID := normalizeJIDFromLID(context.Background(), evt.Info.Chat, c.waClient).String()
	if chatJID == "" {
		return
	}

	// Fase 1: bot tidak boleh membalas story (status@broadcast) maupun channel
	// (@newsletter). Pesan dari JID ini bukan percakapan nyata.
	if isSkippedJID(chatJID) {
		return
	}

	senderJID := evt.Info.Sender.User
	if senderJID == "" {
		senderJID = evt.Info.Chat.User
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
	// Catatan nama field: proto waE2E menamai field call log "CallLogMesssage"
	// (typo tiga huruf s) dan "BcallMessage" (c kecil) — ikuti nama field Go.
	case msg.Call != nil || msg.CallLogMesssage != nil || msg.BcallMessage != nil || msg.ScheduledCallCreationMessage != nil || msg.ScheduledCallEditMessage != nil:
		return "call"
	case msg.ReactionMessage != nil || msg.EncReactionMessage != nil:
		return "reaction"
	case msg.PollCreationMessage != nil || msg.PollUpdateMessage != nil || msg.PollResultSnapshotMessage != nil || msg.PollAddOptionMessage != nil:
		return "poll"
	case msg.ProtocolMessage != nil || msg.PinInChatMessage != nil || msg.KeepInChatMessage != nil:
		return "system"
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

	// Fase 1: normalisasi LID → nomor HP untuk chat JID, supaya chat yang
	// ditulis oleh pesan live konsisten dengan yang ditulis history sync.
	chatJID := normalizeJIDFromLID(context.Background(), evt.Info.Chat, c.waClient).String()
	if chatJID == "" {
		return
	}

	// Fase 1: story (status@broadcast, 0@s.whatsapp.net) dan channel (@newsletter)
	// bukan percakapan nyata — tidak boleh masuk mirror Inbox.
	if isSkippedJID(chatJID) {
		return
	}

	// Fase 2: nama tampil chat di-resolve dari contact store whatsmeow
	// (FullName/FirstName nama tersimpan pengguna > PushName > nomor HP).
	// Untuk grup: displayName dikosongkan — nama grup otoritatif datang dari
	// history sync (conv.GetName()), dan UpsertChat tidak menimpa display_name
	// yang kosong, sehingga nama grup tidak tertimpa nama pengirim.
	var displayName string
	switch {
	case evt.Info.IsGroup:
		displayName = ""
	default:
		displayName = c.resolveChatDisplayName(context.Background(), evt.Info.Chat, evt.Info.PushName)
	}

	content := extractMessageBody(evt.Message)
	mediaType := extractMediaType(evt.Message)

	// Pesan tanpa isi & tanpa media yang dikenali — reaction, protocol
	// (delete/edit/pin), placeholder, envelope internal, pesan tak terdekripsi
	// — bukan pesan percakapan yang bisa dirender. Tanpa guard ini UI dipenuhi
	// bubble "[media]" yang tidak ada di HP (mis. kontak me-react pesan kita
	// muncul seolah "balasan media"). Aturan sama dengan history sync.
	if evt.Message == nil || (content == "" && (mediaType == "unknown" || mediaType == "reaction" || mediaType == "system")) {
		return
	}

	// Fase 4: normalisasi LID → nomor HP agar senderJID tidak tampil sebagai
	// "12345@lid" yang tidak dapat dikenali pengguna.
	senderJID := normalizeJIDFromLID(context.Background(), evt.Info.Sender, c.waClient).String()

	msg := &bot.WAMessage{
		SessionID:   c.SessionID,
		ChatJID:     chatJID,
		WAMessageID: evt.Info.ID,
		SenderJID:   senderJID,
		SenderName:  evt.Info.PushName,
		Content:     content,
		MediaType:   mediaType,
		IsFromMe:    evt.Info.IsFromMe,
		Timestamp:   evt.Info.Timestamp,
	}
	if evt.Info.IsFromMe {
		msg.Status = "sent"
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
		Status:      "sent",
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

// handleChatPresenceEvent memproses events.ChatPresence (typing/recording)
// dan meneruskannya ke ChatPresenceCallback (→ SSE broadcast ke frontend).
//
// WhatsApp hanya mengirim event ini ketika client ditandai online oleh server.
// Referensi: .ref-wa-multidevice/src/infrastructure/whatsapp/event_chat_presence.go
func (c *Client) handleChatPresenceEvent(evt *events.ChatPresence) {
	// Normalisasi LID → nomor HP untuk sender.
	senderJID := normalizeJIDFromLID(context.Background(), evt.Sender, c.waClient)
	chatJID := evt.Chat.ToNonAD().String()
	senderStr := senderJID.ToNonAD().String()

	state := string(evt.State) // "composing" | "paused"
	media := string(evt.Media) // "" | "audio"

	log.Printf("[WhatsApp Client %d] ChatPresence: %s %s in %s (media=%q)",
		c.SessionID, senderStr, state, chatJID, media)

	if cb := c.getChatPresenceCallback(); cb != nil {
		cb(c.SessionID, chatJID, senderStr, state, media, evt.IsGroup)
	}
}

// handleReceiptEvent memproses events.Receipt (acknowledge pengiriman pesan
// keluar) dan memperbarui status centang WhatsApp: "delivered" (✓✓) saat
// pesan sampai di device penerima, "read" (✓✓ biru) saat dibaca.
//
// ReceiptTypeDelivered adalah string kosong ("") di whatsmeow — switch di
// bawah menggunakan konstanta tipe, bukan literal, agar tetap tahan terhadap
// perubahan nilai internal.
func (c *Client) handleReceiptEvent(evt *events.Receipt) {
	if c.chatRepo == nil || evt == nil || len(evt.MessageIDs) == 0 {
		return
	}
	var status string
	switch evt.Type {
	case types.ReceiptTypeDelivered:
		status = "delivered"
	case types.ReceiptTypeRead, types.ReceiptTypeReadSelf:
		status = "read"
	default:
		// receipt sender/retry/played tidak mengubah status centang.
		return
	}
	chatJID := normalizeJIDFromLID(context.Background(), evt.Chat, c.waClient).String()
	if err := c.chatRepo.MarkMessagesStatus(c.SessionID, chatJID, evt.MessageIDs, status); err != nil {
		log.Printf("[WhatsApp Client %d] Failed to mark %d messages as %s in %s: %v",
			c.SessionID, len(evt.MessageIDs), status, chatJID, err)
		return
	}
	log.Printf("[WhatsApp Client %d] Marked %d message(s) as %s in %s", c.SessionID, len(evt.MessageIDs), status, chatJID)
	// Beri tahu UI (via SSE) agar centang di bubble ter-update instan.
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
