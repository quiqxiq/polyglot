package port

import (
	"context"

	"github.com/quixiq/polyglot/internal/domain/bot"
)

// ChatRepository menyimpan mirror percakapan WhatsApp nyata (semua chat dan
// pesan per perangkat), terpisah dari penyimpanan conversation bot.
type ChatRepository interface {
	// UpsertChat menulis atau memperbarui satu chat mirror (per session + chat_jid).
	UpsertChat(ctx context.Context, chat *bot.WAChat) error
	// UpsertMessage menulis satu pesan mirror; idempotent per (session, wa_message_id).
	// Mengembalikan true bila baris baru benar-benar dibuat (untuk menghitung unread
	// tanpa risiko dobel-hitung saat event terkirim ulang).
	UpsertMessage(ctx context.Context, msg *bot.WAMessage) (bool, error)
	// UpsertMessagesBatch menulis banyak pesan mirror dalam SATU pernyataan SQL
	// (multi-row INSERT ... ON CONFLICT DO NOTHING), idempotent per
	// (session, wa_message_id). Dipakai sinkronisasi history sync yang bisa
	// membawa ribuan pesan — memotong ribuan round-trip menjadi beberapa saja.
	// Mengembalikan jumlah baris yang benar-benar baru dibuat.
	UpsertMessagesBatch(ctx context.Context, msgs []*bot.WAMessage) (int, error)
	// IncrementUnread menaikkan hitungan belum-dibaca sebuah chat (pesan masuk).
	IncrementUnread(ctx context.Context, sessionID uint, chatJID string) error
	// MarkChatRead mereset hitungan belum-dibaca sebuah chat.
	MarkChatRead(ctx context.Context, sessionID uint, chatJID string) error
	// SetChatUnread menetapkan hitungan belum-dibaca sebuah chat ke nilai
	// eksplisit (dipakai saat sinkronisasi history sync agar angka unread dari
	// HP tercermin di Inbox).
	SetChatUnread(ctx context.Context, sessionID uint, chatJID string, count uint32) error
	// ListChats mengembalikan daftar chat mirror, diurutkan dari paling baru.
	ListChats(ctx context.Context, sessionID uint, limit, offset int, search string) ([]bot.WAChat, error)
	// ListChatMessages mengembalikan pesan sebuah chat, diurutkan ascending.
	ListChatMessages(ctx context.Context, sessionID uint, chatJID string, limit, offset int) ([]bot.WAMessage, error)
	// MarkMessagesStatus memperbarui status pengiriman pesan keluar
	// ("sent" → "delivered" → "read") untuk message IDs tertentu. Dipanggil
	// dari handler events.Receipt whatsmeow. Pesan yang statusnya "read"
	// otomatis ditandai is_read = true.
	MarkMessagesStatus(ctx context.Context, sessionID uint, chatJID string, messageIDs []string, status string) error
	// MergeChatLID menggabungkan baris chat @lid basi ke baris nomor HP-nya:
	// pesan @lid dipindah ke PN (abaikan yang sudah ada), lalu baris chat @lid
	// dihapus. Dipakai agar chat hasil sync sebelum LID map tersedia tidak
	// mendobel / tampil sebagai nomor LID yang tidak dikenal di Inbox.
	MergeChatLID(ctx context.Context, sessionID uint, lidJID, pnJID string) error
	// SetChatBotEnabled mengubah status bot auto-reply per chat.
	SetChatBotEnabled(ctx context.Context, sessionID uint, chatJID string, enabled bool) error
	// IsChatBotEnabled melaporkan status bot auto-reply per chat. Chat yang
	// belum pernah tercatat dianggap aktif (default TRUE).
	IsChatBotEnabled(ctx context.Context, sessionID uint, chatJID string) (bool, error)
}

