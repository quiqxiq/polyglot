package postgres

import (
	"context"
	"errors"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/quixiq/polyglot/internal/adapter/postgres/model"
	"github.com/quixiq/polyglot/internal/domain/bot"
	"github.com/quixiq/polyglot/internal/port"
)

const (
	defaultChatListLimit = 50
	maxChatListLimit     = 200
)

// ChatRepository implements port.ChatRepository for GORM/Postgres.
type ChatRepository struct {
	db *gorm.DB
}

var _ port.ChatRepository = (*ChatRepository)(nil)

// NewChatRepository creates a new ChatRepository.
func NewChatRepository(db *gorm.DB) *ChatRepository {
	return &ChatRepository{db: db}
}

// UpsertChat writes or updates one WhatsApp chat mirror row.
func (r *ChatRepository) UpsertChat(ctx context.Context, chat *bot.WAChat) error {
	m := model.WAChatModelFromDomain(chat)
	if m == nil {
		return nil
	}
	updates := map[string]any{
		"is_group":             m.IsGroup,
		"last_message_id":      m.LastMessageID,
		"last_message_preview": m.LastMessagePreview,
		"last_message_time":    m.LastMessageTime,
		"updated_at":           m.UpdatedAt,
	}
	if m.DisplayName != "" {
		updates["display_name"] = m.DisplayName
	}
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "session_id"}, {Name: "chat_jid"}},
		DoUpdates: clause.Assignments(updates),
	}).Create(m).Error
}

// UpsertMessage writes one WhatsApp message mirror row.
func (r *ChatRepository) UpsertMessage(ctx context.Context, msg *bot.WAMessage) (bool, error) {
	m := model.WAMessageModelFromDomain(msg)
	if m == nil {
		return false, nil
	}
	res := r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "session_id"}, {Name: "wa_message_id"}},
		DoNothing: true,
	}).Create(m)
	return res.RowsAffected > 0, res.Error
}

// UpsertMessagesBatch writes many message mirror rows in a single multi-row
// INSERT ... ON CONFLICT DO NOTHING statement.
func (r *ChatRepository) UpsertMessagesBatch(ctx context.Context, msgs []*bot.WAMessage) (int, error) {
	if len(msgs) == 0 {
		return 0, nil
	}
	mList := make([]model.WAMessageModel, 0, len(msgs))
	for _, m := range msgs {
		if mm := model.WAMessageModelFromDomain(m); mm != nil {
			mList = append(mList, *mm)
		}
	}
	if len(mList) == 0 {
		return 0, nil
	}
	res := r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "session_id"}, {Name: "wa_message_id"}},
		DoNothing: true,
	}).Create(&mList)
	if res.Error != nil {
		inserted := 0
		for _, m := range msgs {
			ok, err := r.UpsertMessage(ctx, m)
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
func (r *ChatRepository) IncrementUnread(ctx context.Context, sessionID uint, chatJID string) error {
	return r.db.WithContext(ctx).Model(&model.WAChatModel{}).
		Where("session_id = ? AND chat_jid = ?", sessionID, chatJID).
		UpdateColumn("unread_count", gorm.Expr("unread_count + 1")).Error
}

// MarkChatRead resets the unread counter of one chat.
func (r *ChatRepository) MarkChatRead(ctx context.Context, sessionID uint, chatJID string) error {
	return r.db.WithContext(ctx).Model(&model.WAChatModel{}).
		Where("session_id = ? AND chat_jid = ?", sessionID, chatJID).
		UpdateColumn("unread_count", 0).Error
}

// SetChatUnread sets the unread counter of one chat to an exact value.
func (r *ChatRepository) SetChatUnread(ctx context.Context, sessionID uint, chatJID string, count uint32) error {
	return r.db.WithContext(ctx).Model(&model.WAChatModel{}).
		Where("session_id = ? AND chat_jid = ?", sessionID, chatJID).
		UpdateColumn("unread_count", count).Error
}

// MarkMessagesStatus memperbarui status pengiriman pesan keluar.
func (r *ChatRepository) MarkMessagesStatus(ctx context.Context, sessionID uint, chatJID string, messageIDs []string, status string) error {
	if len(messageIDs) == 0 {
		return nil
	}
	updates := map[string]any{"status": status}
	if status == "read" {
		updates["is_read"] = true
	}
	return r.db.WithContext(ctx).Model(&model.WAMessageModel{}).
		Where("session_id = ? AND chat_jid = ? AND wa_message_id IN ?", sessionID, chatJID, messageIDs).
		Updates(updates).Error
}

// MergeChatLID menggabungkan baris chat @lid basi ke baris nomor HP-nya dalam satu transaksi.
func (r *ChatRepository) MergeChatLID(ctx context.Context, sessionID uint, lidJID, pnJID string) error {
	if lidJID == "" || pnJID == "" || lidJID == pnJID {
		return nil
	}
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(`
			INSERT INTO wa_messages
				(session_id, chat_jid, wa_message_id, sender_jid, sender_name, content,
				 media_type, is_from_me, is_read, status, timestamp, created_at)
			SELECT session_id, ?, wa_message_id, sender_jid, sender_name, content,
			       media_type, is_from_me, is_read, status, timestamp, COALESCE(created_at, now())
			FROM wa_messages
			WHERE session_id = ? AND chat_jid = ?
			ON CONFLICT (session_id, wa_message_id) DO NOTHING`,
			pnJID, sessionID, lidJID).Error; err != nil {
			return err
		}
		if err := tx.Exec(`DELETE FROM wa_messages WHERE session_id = ? AND chat_jid = ?`, sessionID, lidJID).Error; err != nil {
			return err
		}
		return tx.Exec(`DELETE FROM wa_chats WHERE session_id = ? AND chat_jid = ?`, sessionID, lidJID).Error
	})
}

// ListChats returns the chat mirror list ordered by most recent activity.
func (r *ChatRepository) ListChats(ctx context.Context, sessionID uint, limit, offset int, search string) ([]bot.WAChat, error) {
	if limit <= 0 || limit > maxChatListLimit {
		limit = defaultChatListLimit
	}

	q := r.db.WithContext(ctx).Where("session_id = ?", sessionID)
	q = q.Where("chat_jid NOT IN (?, ?) AND chat_jid NOT LIKE ?",
		"status@broadcast", "0@s.whatsapp.net", "%@newsletter")
	search = strings.TrimSpace(search)
	if search != "" {
		like := "%" + search + "%"
		if r.db.Dialector.Name() == "postgres" {
			q = q.Where("display_name ILIKE ? OR chat_jid ILIKE ?", like, like)
		} else {
			q = q.Where("display_name LIKE ? OR chat_jid LIKE ?", like, like)
		}
	}

	var mList []model.WAChatModel
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
func (r *ChatRepository) SetChatBotEnabled(ctx context.Context, sessionID uint, chatJID string, enabled bool) error {
	return r.db.WithContext(ctx).Model(&model.WAChatModel{}).
		Where("session_id = ? AND chat_jid = ?", sessionID, chatJID).
		Update("bot_enabled", enabled).Error
}

// IsChatBotEnabled reports the per-chat bot auto-reply flag.
func (r *ChatRepository) IsChatBotEnabled(ctx context.Context, sessionID uint, chatJID string) (bool, error) {
	var m model.WAChatModel
	err := r.db.WithContext(ctx).Select("bot_enabled").
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
func (r *ChatRepository) ListChatMessages(ctx context.Context, sessionID uint, chatJID string, limit, offset int) ([]bot.WAMessage, error) {
	if limit <= 0 || limit > maxChatListLimit {
		limit = defaultChatListLimit
	}

	var mList []model.WAMessageModel
	if err := r.db.WithContext(ctx).Where("session_id = ? AND chat_jid = ?", sessionID, chatJID).
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

// ─── Backward compatibility delegations on Store ──────────────────────────────

func (s *Store) UpsertChat(ctx context.Context, chat *bot.WAChat) error {
	return NewChatRepository(s.db).UpsertChat(ctx, chat)
}

func (s *Store) UpsertMessage(ctx context.Context, msg *bot.WAMessage) (bool, error) {
	return NewChatRepository(s.db).UpsertMessage(ctx, msg)
}

func (s *Store) UpsertMessagesBatch(ctx context.Context, msgs []*bot.WAMessage) (int, error) {
	return NewChatRepository(s.db).UpsertMessagesBatch(ctx, msgs)
}

func (s *Store) IncrementUnread(ctx context.Context, sessionID uint, chatJID string) error {
	return NewChatRepository(s.db).IncrementUnread(ctx, sessionID, chatJID)
}

func (s *Store) MarkChatRead(ctx context.Context, sessionID uint, chatJID string) error {
	return NewChatRepository(s.db).MarkChatRead(ctx, sessionID, chatJID)
}

func (s *Store) SetChatUnread(ctx context.Context, sessionID uint, chatJID string, count uint32) error {
	return NewChatRepository(s.db).SetChatUnread(ctx, sessionID, chatJID, count)
}

func (s *Store) MarkMessagesStatus(ctx context.Context, sessionID uint, chatJID string, messageIDs []string, status string) error {
	return NewChatRepository(s.db).MarkMessagesStatus(ctx, sessionID, chatJID, messageIDs, status)
}

func (s *Store) MergeChatLID(ctx context.Context, sessionID uint, lidJID, pnJID string) error {
	return NewChatRepository(s.db).MergeChatLID(ctx, sessionID, lidJID, pnJID)
}

func (s *Store) ListChats(ctx context.Context, sessionID uint, limit, offset int, search string) ([]bot.WAChat, error) {
	return NewChatRepository(s.db).ListChats(ctx, sessionID, limit, offset, search)
}

func (s *Store) SetChatBotEnabled(ctx context.Context, sessionID uint, chatJID string, enabled bool) error {
	return NewChatRepository(s.db).SetChatBotEnabled(ctx, sessionID, chatJID, enabled)
}

func (s *Store) IsChatBotEnabled(ctx context.Context, sessionID uint, chatJID string) (bool, error) {
	return NewChatRepository(s.db).IsChatBotEnabled(ctx, sessionID, chatJID)
}

func (s *Store) ListChatMessages(ctx context.Context, sessionID uint, chatJID string, limit, offset int) ([]bot.WAMessage, error) {
	return NewChatRepository(s.db).ListChatMessages(ctx, sessionID, chatJID, limit, offset)
}
