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
}

var _ port.WhatsAppGateway = (*SessionManager)(nil)

func NewSessionManager(postgresConnStr string, onMsg MessageCallback, onStat StatusCallback) (*SessionManager, error) {
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

	client := NewClient(session.ID, deviceStore, sm.onMessage, sm.onStatus)
	sm.clients[session.ID] = client

	return client.Connect(ctx)
}

func (sm *SessionManager) Connect(session *bot.WASession) error {
	return sm.ConnectWithContext(context.Background(), session)
}

func (sm *SessionManager) Disconnect(sessionID uint) error {
	if sm == nil {
		return nil
	}
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	if client, ok := sm.clients[sessionID]; ok {
		client.Disconnect()
		if client.deviceStore != nil {
			_ = client.deviceStore.Delete(context.Background())
		}
		delete(sm.clients, sessionID)
	}
	return nil
}

func (sm *SessionManager) Logout(sessionID uint) error {
	if sm == nil {
		return nil
	}
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	if client, ok := sm.clients[sessionID]; ok {
		_ = client.Logout(context.Background())
		if client.deviceStore != nil {
			_ = client.deviceStore.Delete(context.Background())
		}
		delete(sm.clients, sessionID)
	}
	return nil
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
	if !ok || !client.waClient.IsConnected() {
		return "offline", nil
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
	return client.GetQRCode(), nil
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
		if sess.Status == bot.StatusOnline || sess.IsBotEnabled {
			if err := sm.Connect(sess); err != nil {
				log.Printf("[SessionManager] Warning: Failed to connect session %d (%s): %v", sess.ID, sess.DeviceName, err)
			}
		}
	}
	return nil
}
