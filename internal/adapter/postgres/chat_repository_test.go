package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/quixiq/polyglot/internal/adapter/postgres"
	"github.com/quixiq/polyglot/internal/adapter/postgres/model"
	"github.com/quixiq/polyglot/internal/domain/bot"
)

func setupChatTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(
		&model.WAChatModel{},
		&model.WAMessageModel{},
	)
	require.NoError(t, err)

	return db
}

func TestChatRepository_UpsertAndList(t *testing.T) {
	db := setupChatTestDB(t)
	repo := postgres.NewChatRepository(db)
	ctx := context.Background()

	now := time.Now().UTC()
	chat := &bot.WAChat{
		SessionID:          1,
		ChatJID:            "628111222333@s.whatsapp.net",
		DisplayName:        "Pelanggan VIP",
		IsGroup:            false,
		LastMessagePreview: "Terima kasih infonya",
		LastMessageTime:    now,
	}

	// 1. UpsertChat
	err := repo.UpsertChat(ctx, chat)
	require.NoError(t, err)

	// 2. ListChats
	chats, err := repo.ListChats(ctx, 1, 10, 0, "")
	require.NoError(t, err)
	assert.Len(t, chats, 1)
	assert.Equal(t, "Pelanggan VIP", chats[0].DisplayName)

	// 3. UpsertMessage
	msg := &bot.WAMessage{
		SessionID:   1,
		ChatJID:     "628111222333@s.whatsapp.net",
		WAMessageID: "WAMID-001",
		SenderJID:   "628111222333@s.whatsapp.net",
		Content:     "Terima kasih infonya",
		Timestamp:   now,
	}
	isNew, err := repo.UpsertMessage(ctx, msg)
	require.NoError(t, err)
	assert.True(t, isNew)

	// 4. Batch Messages
	msg2 := &bot.WAMessage{
		SessionID:   1,
		ChatJID:     "628111222333@s.whatsapp.net",
		WAMessageID: "WAMID-002",
		SenderJID:   "628111222333@s.whatsapp.net",
		Content:     "Pesan kedua",
		Timestamp:   now.Add(time.Second),
	}
	inserted, err := repo.UpsertMessagesBatch(ctx, []*bot.WAMessage{msg2})
	require.NoError(t, err)
	assert.Equal(t, 1, inserted)

	// 5. ListChatMessages
	msgs, err := repo.ListChatMessages(ctx, 1, "628111222333@s.whatsapp.net", 10, 0)
	require.NoError(t, err)
	assert.Len(t, msgs, 2)

	// 6. Unread counter operations
	err = repo.IncrementUnread(ctx, 1, "628111222333@s.whatsapp.net")
	require.NoError(t, err)

	err = repo.SetChatUnread(ctx, 1, "628111222333@s.whatsapp.net", 5)
	require.NoError(t, err)

	err = repo.MarkChatRead(ctx, 1, "628111222333@s.whatsapp.net")
	require.NoError(t, err)

	// 7. Mark Messages Status
	err = repo.MarkMessagesStatus(ctx, 1, "628111222333@s.whatsapp.net", []string{"WAMID-001"}, "read")
	require.NoError(t, err)

	// 8. Bot Enabled Toggle & Check
	err = repo.SetChatBotEnabled(ctx, 1, "628111222333@s.whatsapp.net", false)
	require.NoError(t, err)

	enabled, err := repo.IsChatBotEnabled(ctx, 1, "628111222333@s.whatsapp.net")
	require.NoError(t, err)
	assert.False(t, enabled)
}
