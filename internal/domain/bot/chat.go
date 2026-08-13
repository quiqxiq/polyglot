package bot

import "time"

// WAChat adalah mirror satu percakapan WhatsApp nyata pada satu perangkat
// (session). Berbeda dengan Conversation (yang hanya dibuat saat pelanggan
// menghubungi bot), WAChat mencatat SEMUA chat sehingga Inbox dapat
// menampilkan seluruh percakapan seperti aplikasi WhatsApp.
type WAChat struct {
	ID                 uint      `json:"id"`
	SessionID          uint      `json:"session_id"`
	ChatJID            string    `json:"chat_jid"`
	DisplayName        string    `json:"display_name"`
	IsGroup            bool      `json:"is_group"`
	LastMessageID      string    `json:"last_message_id"`
	LastMessagePreview string    `json:"last_message_preview"`
	LastMessageTime    time.Time `json:"last_message_time"`
	UnreadCount        int       `json:"unread_count"`
	// BotEnabled menandakan apakah bot auto-reply aktif untuk chat ini
	// (per-chat control, independen dari switch bot perangkat).
	BotEnabled bool      `json:"bot_enabled"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// WAMessage adalah mirror satu pesan WhatsApp nyata pada satu perangkat.
// Status mengikuti konvensi WhatsApp: "sent" (✓), "delivered" (✓✓), "read"
// (✓✓ biru) — dipakai frontend untuk merender centang pada pesan keluar.
type WAMessage struct {
	ID          uint      `json:"id"`
	SessionID   uint      `json:"session_id"`
	ChatJID     string    `json:"chat_jid"`
	WAMessageID string    `json:"wa_message_id"`
	SenderJID   string    `json:"sender_jid"`
	SenderName  string    `json:"sender_name"`
	Content     string    `json:"content"`
	MediaType   string    `json:"media_type"`
	IsFromMe    bool      `json:"is_from_me"`
	IsRead      bool      `json:"is_read"`
	Status      string    `json:"status"`
	Timestamp   time.Time `json:"timestamp"`
	CreatedAt   time.Time `json:"created_at"`
}
