package whatsapp

import (
	"context"
	"fmt"
	"log"
	"sync"

	_ "github.com/lib/pq"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/store/sqlstore"
	waLog "go.mau.fi/whatsmeow/util/log"

	"github.com/quixiq/polyglot/internal/domain/bot"
	"github.com/quixiq/polyglot/internal/port"
)

type SessionManager struct {
	container *sqlstore.Container
	clients   map[uint]*Client
	mutex     sync.RWMutex
	onMessage MessageCallback
	onStatus  StatusCallback
	onChatUpd ChatUpdateCallback
	chatRepo  port.ChatRepository
}

var _ port.WhatsAppGateway = (*SessionManager)(nil)

func NewSessionManager(postgresConnStr string, chatRepo port.ChatRepository, onMsg MessageCallback, onStat StatusCallback) (*SessionManager, error) {
	dbLogger := waLog.Stdout("Database", "INFO", true)
	ctx := context.Background()
	container, err := sqlstore.New(ctx, "postgres", postgresConnStr, dbLogger)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize whatsmeow postgres sqlstore: %w", err)
	}

	return &SessionManager{
		container: container,
		clients:   make(map[uint]*Client),
		onMessage: onMsg,
		onStatus:  onStat,
		chatRepo:  chatRepo,
	}, nil
}

func (sm *SessionManager) ConnectWithContext(ctx context.Context, session *bot.WASession) error {
	if sm == nil {
		return nil
	}
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	if client, ok := sm.clients[session.ID]; ok && client.waClient.IsConnected() {
		return nil
	}

	devices, err := sm.container.GetAllDevices(ctx)
	if err != nil {
		return fmt.Errorf("failed to get device stores: %w", err)
	}

	var deviceStore *store.Device

	for _, d := range devices {
		if d.ID != nil {
			if session.JID != "" && (d.ID.String() == session.JID || d.ID.User == session.JID) {
				deviceStore = d
				break
			}
			if session.PhoneNumber != "" && d.ID.User == session.PhoneNumber {
				deviceStore = d
				break
			}
		}
	}

	if deviceStore == nil && len(devices) > 0 {
		for _, d := range devices {
			if d.ID != nil {
				deviceStore = d
				break
			}
		}
	}

	if deviceStore == nil {
		for _, d := range devices {
			if d.ID == nil {
				deviceStore = d
				break
			}
		}
	}

	if deviceStore == nil {
		deviceStore = sm.container.NewDevice()
	}

	client := NewClient(session.ID, deviceStore, sm.onMessage, sm.onStatus, sm.chatRepo)
	// Callback chat-update diteruskan ke client yang BARU dibuat (lock sudah
	// dipegang di fungsi ini, baca sm.onChatUpd aman). Manager-level
	// SetChatUpdateCallback hanya menjangkau client yang sudah terdaftar,
	// sehingga tanpa ini client baru punya onChatUpd nil dan SSE chat_update
	// tidak pernah terkirim untuk session tersebut.
	client.SetChatUpdateCallback(sm.onChatUpd)
	sm.clients[session.ID] = client

	return client.Connect(ctx)
}

func (sm *SessionManager) Connect(session *bot.WASession) error {
	return sm.ConnectWithContext(context.Background(), session)
}

// Disconnect hanya memutus koneksi — session lokal & pairing DIpertahankan
// sehingga Reconnect bisa menyambung lagi tanpa scan ulang. Berbeda dengan
// Logout yang menghapus session dari store.
func (sm *SessionManager) Disconnect(sessionID uint) error {
	if sm == nil {
		return nil
	}
	sm.mutex.RLock()
	client, ok := sm.clients[sessionID]
	sm.mutex.RUnlock()

	if !ok {
		return nil
	}
	client.Disconnect()
	return nil
}

// Logout unlink device dari WhatsApp (best-effort), menghapus session lokal
// dari store whatsmeow, dan melepas client dari registry in-memory. Baris
// session di DB DIpertahankan (keep-slot) agar slot tetap tampil di UI dan
// bisa di-pair ulang dengan QR baru di bawah id yang sama.
func (sm *SessionManager) Logout(sessionID uint) error {
	if sm == nil {
		return nil
	}
	sm.mutex.Lock()
	client, ok := sm.clients[sessionID]
	if ok {
		_ = client.Logout(context.Background())
		delete(sm.clients, sessionID)
	}
	sm.mutex.Unlock()

	// Slot DB tetap; beri tahu UI bahwa device perlu re-pair.
	if sm.onStatus != nil {
		sm.onStatus(sessionID, "needs_rescan", "", "", "")
	}
	return nil
}

// Purge sama dengan Logout pada sisi gateway (unlink + hapus session lokal +
// lepas dari registry). Penghapusan baris DB dan mirror chat (wa_chats,
// wa_messages, conversations) dilakukan oleh caller (handler) yang punya
// akses repository.
func (sm *SessionManager) Purge(sessionID uint) error {
	return sm.Logout(sessionID)
}

func (sm *SessionManager) Reconnect(sessionID uint) error {
	if sm == nil {
		return nil
	}
	sm.mutex.Lock()
	client, ok := sm.clients[sessionID]
	sm.mutex.Unlock()

	if !ok {
		sess := &bot.WASession{ID: sessionID}
		return sm.Connect(sess)
	}
	return client.Reconnect(context.Background())
}

func (sm *SessionManager) SendMessage(sessionID uint, to string, content string) error {
	return sm.SendMessageContext(context.Background(), sessionID, to, content)
}

func (sm *SessionManager) SendMessageContext(ctx context.Context, sessionID uint, to string, content string) error {
	if sm == nil {
		return fmt.Errorf("session manager not initialized")
	}
	sm.mutex.RLock()
	client, ok := sm.clients[sessionID]
	sm.mutex.RUnlock()

	if !ok {
		return fmt.Errorf("session %d is not connected or registered", sessionID)
	}

	return client.SendMessage(ctx, to, content)
}

func (sm *SessionManager) SendDocument(ctx context.Context, sessionID uint, to string, fileBytes []byte, fileName string, contentType string, caption string) error {
	if sm == nil {
		return fmt.Errorf("session manager not initialized")
	}
	sm.mutex.RLock()
	client, ok := sm.clients[sessionID]
	sm.mutex.RUnlock()

	if !ok {
		return fmt.Errorf("session %d is not connected or registered", sessionID)
	}

	return client.SendDocument(ctx, to, fileBytes, fileName, contentType, caption)
}

func (sm *SessionManager) SendImage(ctx context.Context, sessionID uint, to string, imageBytes []byte, contentType string, caption string) error {
	if sm == nil {
		return fmt.Errorf("session manager not initialized")
	}
	sm.mutex.RLock()
	client, ok := sm.clients[sessionID]
	sm.mutex.RUnlock()

	if !ok {
		return fmt.Errorf("session %d is not connected or registered", sessionID)
	}

	return client.SendImage(ctx, to, imageBytes, contentType, caption)
}

func (sm *SessionManager) GetStatus(sessionID uint) (string, error) {
	if sm == nil {
		return "offline", nil
	}
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	client, ok := sm.clients[sessionID]
	if !ok {
		return "offline", nil
	}
	if !client.waClient.IsConnected() {
		return "offline", nil
	}
	if client.waClient.Store.ID == nil {
		return "needs_rescan", nil
	}
	return "online", nil
}

func (sm *SessionManager) GetQRCode(sessionID uint) (string, error) {
	if sm == nil {
		return "", nil
	}
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	client, ok := sm.clients[sessionID]
	if !ok {
		return "", fmt.Errorf("session %d not found", sessionID)
	}

	qr := client.GetQRCode()
	if qr == "" && client.waClient.Store.ID == nil {
		// Belum ter-pair dan QR kosong (baru dibuat / QR timeout) — restart
		// aliran QR supaya polling berikutnya mendapat QR baru tanpa perlu
		// tombol Reconnect manual. EnsureQRFlow idempoten (guard internal).
		client.EnsureQRFlow()
		qr = client.GetQRCode()
	}
	return qr, nil
}

func (sm *SessionManager) GetPairingCode(sessionID uint, phoneNumber string) (string, error) {
	if sm == nil {
		return "", nil
	}
	sm.mutex.RLock()
	client, ok := sm.clients[sessionID]
	sm.mutex.RUnlock()

	if !ok {
		return "", fmt.Errorf("session %d not found", sessionID)
	}
	ctx := context.Background()
	return client.GetPairingCode(ctx, phoneNumber)
}

func (sm *SessionManager) RestoreAllSessions(sessions []bot.WASession) error {
	if sm == nil {
		return nil
	}
	log.Printf("[SessionManager] Restoring %d WhatsApp sessions...", len(sessions))
	for i := range sessions {
		sess := &sessions[i]
		// Hanya session yang pernah online/logged-in yang di-restore otomatis.
		// Session needs_rescan (belum pernah paired / sudah logout) dibiarkan
		// menunggu scan manual dari UI.
		if sess.Status == bot.StatusOnline || sess.JID != "" {
			if err := sm.Connect(sess); err != nil {
				log.Printf("[SessionManager] Warning: Failed to connect session %d (%s): %v", sess.ID, sess.DeviceName, err)
			}
		}
	}
	return nil
}

// SetMessageCallback registers (or replaces) the incoming-message handler for
// all clients — termasuk yang sudah terhubung. Dipanggil setelah Engine bot
// dibangun untuk memutus circular dependency SessionManager <-> Engine.
func (sm *SessionManager) SetMessageCallback(cb MessageCallback) {
	if sm == nil {
		return
	}
	// Mutasi state (sm.onMessage) — harus write lock, bukan RLock.
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	sm.onMessage = cb
	for _, c := range sm.clients {
		c.SetMessageCallback(cb)
	}
}

// SetChatUpdateCallback registers (or replaces) the chat-mirror-change handler
// for all clients — termasuk yang sudah terhubung. Dipakai untuk broadcast
// SSE `chat_update` dari EventHandler (pemilik SSE hub).
func (sm *SessionManager) SetChatUpdateCallback(cb ChatUpdateCallback) {
	if sm == nil {
		return
	}
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	sm.onChatUpd = cb
	for _, c := range sm.clients {
		c.SetChatUpdateCallback(cb)
	}
}
