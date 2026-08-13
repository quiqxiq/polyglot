package postgres

import (
	"errors"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/quixiq/polyglot/internal/adapter/postgres/models"
	"github.com/quixiq/polyglot/internal/domain/bot"
)

const (
	defaultChatListLimit = 50
	maxChatListLimit     = 200
)

// UpsertChat writes or updates one WhatsApp chat mirror row.
// Catatan: kolom conflict-update TIDAK menyertakan bot_enabled — toggle bot
// per-chat milik agen tidak boleh ter-reset oleh penulisan pesan berikutnya.
func (s *Store) UpsertChat(chat *bot.WAChat) error {
	m := models.WAChatModelFromDomain(chat)
	if m == nil {
		return nil
	}
	return s.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "session_id"}, {Name: "chat_jid"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"display_name", "is_group", "last_message_id", "last_message_preview",
			"last_message_time", "updated_at",
		}),
	}).Create(m).Error
}

// UpsertMessage writes one WhatsApp message mirror row. The write is idempotent
// per (session_id, wa_message_id): when the row already exists it is left
// untouched (DoNothing) and the boolean result reports that nothing new was
// inserted — callers use that to avoid double-counting unread on replayed events.
func (s *Store) UpsertMessage(msg *bot.WAMessage) (bool, error) {
	m := models.WAMessageModelFromDomain(msg)
	if m == nil {
		return false, nil
	}
	res := s.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "session_id"}, {Name: "wa_message_id"}},
		DoNothing: true,
	}).Create(m)
	return res.RowsAffected > 0, res.Error
}

// UpsertMessagesBatch writes many message mirror rows in a single multi-row
// INSERT ... ON CONFLICT DO NOTHING statement. Idempotent per
// (session_id, wa_message_id) — rows that already exist are left untouched.
// Returns the number of rows actually inserted. Dipakai oleh sinkronisasi
// history sync yang bisa membawa ribuan pesan per blob.
func (s *Store) UpsertMessagesBatch(msgs []*bot.WAMessage) (int, error) {
	if len(msgs) == 0 {
		return 0, nil
	}
	mList := make([]models.WAMessageModel, 0, len(msgs))
	for _, m := range msgs {
		if mm := models.WAMessageModelFromDomain(m); mm != nil {
			mList = append(mList, *mm)
		}
	}
	if len(mList) == 0 {
		return 0, nil
	}
	res := s.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "session_id"}, {Name: "wa_message_id"}},
		DoNothing: true,
	}).Create(&mList)
	if res.Error != nil {
		// PostgreSQL meng-abort SELURUH statement multi-row jika SATU baris
		// melanggar constraint (mis. content melebihi panjang kolom). Fallback
		// per-baris mengembalikan perilaku lama (satu pesan bermasalah hanya
		// menjatuhkan pesan itu, bukan seluruh chunk). Jalur ini jarang —
		// hanya dieksekusi saat statement batch gagal.
		inserted := 0
		for _, m := range msgs {
			ok, err := s.UpsertMessage(m)
			if err != nil {
				return inserted, err
			}
			if ok {
				inserted++
			}
		}
		return inserted, nil
	}
	return int(res.RowsAffected), res.Error
}

// IncrementUnread bumps the unread counter of one chat (incoming message).
func (s *Store) IncrementUnread(sessionID uint, chatJID string) error {
	return s.db.Model(&models.WAChatModel{}).
		Where("session_id = ? AND chat_jid = ?", sessionID, chatJID).
		UpdateColumn("unread_count", gorm.Expr("unread_count + 1")).Error
}

// MarkChatRead resets the unread counter of one chat.
func (s *Store) MarkChatRead(sessionID uint, chatJID string) error {
	return s.db.Model(&models.WAChatModel{}).
		Where("session_id = ? AND chat_jid = ?", sessionID, chatJID).
		UpdateColumn("unread_count", 0).Error
}

// SetChatUnread sets the unread counter of one chat to an exact value. Dipakai
// saat sinkronisasi history sync agar angka unread dari perangkat tercermin
// di Inbox (IncrementUnread hanya bisa menaikkan, tidak menetapkan).
func (s *Store) SetChatUnread(sessionID uint, chatJID string, count uint32) error {
	return s.db.Model(&models.WAChatModel{}).
		Where("session_id = ? AND chat_jid = ?", sessionID, chatJID).
		UpdateColumn("unread_count", count).Error
}

// ListChats returns the chat mirror list ordered by most recent activity.
func (s *Store) ListChats(sessionID uint, limit, offset int, search string) ([]bot.WAChat, error) {
	if limit <= 0 || limit > maxChatListLimit {
		limit = defaultChatListLimit
	}

	q := s.db.Where("session_id = ?", sessionID)
	search = strings.TrimSpace(search)
	if search != "" {
		like := "%" + search + "%"
		q = q.Where("display_name ILIKE ? OR chat_jid ILIKE ?", like, like)
	}

	var mList []models.WAChatModel
	if err := q.Order("last_message_time DESC NULLS LAST").Limit(limit).Offset(offset).Find(&mList).Error; err != nil {
		return nil, err
	}

	res := make([]bot.WAChat, 0, len(mList))
	for _, m := range mList {
		if d := m.ToDomain(); d != nil {
			res = append(res, *d)
		}
	}
	return res, nil
}

// SetChatBotEnabled toggles the per-chat bot auto-reply flag.
func (s *Store) SetChatBotEnabled(sessionID uint, chatJID string, enabled bool) error {
	return s.db.Model(&models.WAChatModel{}).
		Where("session_id = ? AND chat_jid = ?", sessionID, chatJID).
		Update("bot_enabled", enabled).Error
}

// IsChatBotEnabled reports the per-chat bot auto-reply flag. Chats that do
// not exist yet (no message mirrored) default to enabled.
func (s *Store) IsChatBotEnabled(sessionID uint, chatJID string) (bool, error) {
	var m models.WAChatModel
	err := s.db.Select("bot_enabled").
		Where("session_id = ? AND chat_jid = ?", sessionID, chatJID).
		First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return true, nil
		}
		return true, err
	}
	return m.BotEnabled, nil
}

// ListChatMessages returns the messages of one chat in ascending order.
func (s *Store) ListChatMessages(sessionID uint, chatJID string, limit, offset int) ([]bot.WAMessage, error) {
	if limit <= 0 || limit > maxChatListLimit {
		limit = defaultChatListLimit
	}

	var mList []models.WAMessageModel
	if err := s.db.Where("session_id = ? AND chat_jid = ?", sessionID, chatJID).
		Order("timestamp ASC").Limit(limit).Offset(offset).Find(&mList).Error; err != nil {
		return nil, err
	}

	res := make([]bot.WAMessage, 0, len(mList))
	for _, m := range mList {
		if d := m.ToDomain(); d != nil {
			res = append(res, *d)
		}
	}
	return res, nil
}
