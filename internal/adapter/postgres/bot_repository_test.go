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

func setupBotTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(
		&model.WASessionModel{},
		&model.ConversationModel{},
		&model.MessageModel{},
	)
	require.NoError(t, err)

	return db
}

func TestWASessionRepository_CRUD(t *testing.T) {
	db := setupBotTestDB(t)
	repo := postgres.NewWASessionRepository(db)
	ctx := context.Background()

	// 1. Create Session
	session := &bot.WASession{
		JID:        "628123456789@s.whatsapp.net",
		DeviceName: "Gateway 1",
		Status:     bot.StatusOnline,
	}
	err := repo.CreateSession(ctx, session)
	require.NoError(t, err)
	assert.NotZero(t, session.ID)

	// 2. FindSessionByID
	found, err := repo.FindSessionByID(ctx, session.ID)
	require.NoError(t, err)
	assert.Equal(t, "628123456789@s.whatsapp.net", found.JID)
	assert.Equal(t, bot.StatusOnline, found.Status)

	// 3. FindAllSessions
	all, err := repo.FindAllSessions(ctx)
	require.NoError(t, err)
	assert.Len(t, all, 1)

	// 4. UpdateSession
	session.Status = bot.StatusOffline
	err = repo.UpdateSession(ctx, session)
	require.NoError(t, err)

	afterUp, err := repo.FindSessionByID(ctx, session.ID)
	require.NoError(t, err)
	assert.Equal(t, bot.StatusOffline, afterUp.Status)

	// 5. DeleteSession
	err = repo.DeleteSession(ctx, session.ID)
	require.NoError(t, err)

	_, err = repo.FindSessionByID(ctx, session.ID)
	assert.Error(t, err)
}

func TestConversationRepository_CRUD(t *testing.T) {
	db := setupBotTestDB(t)
	repo := postgres.NewConversationRepository(db)
	ctx := context.Background()

	// 1. Create Conversation
	conv := &bot.Conversation{
		SessionID:        1,
		CustomerWANumber: "628987654321",
		Status:           bot.StatusBot,
	}
	err := repo.CreateConversation(ctx, conv)
	require.NoError(t, err)
	assert.NotZero(t, conv.ID)

	// 2. FindConversationByID
	found, err := repo.FindConversationByID(ctx, conv.ID)
	require.NoError(t, err)
	assert.Equal(t, "628987654321", found.CustomerWANumber)
	assert.Equal(t, bot.StatusBot, found.Status)

	// 3. Create Messages
	msg1 := &bot.Message{
		ConversationID: conv.ID,
		SenderType:     "user",
		Content:        "Halo admin, mau tanya tagihan",
		CreatedAt:      time.Now().UTC(),
	}
	err = repo.CreateMessage(ctx, msg1)
	require.NoError(t, err)
	assert.NotZero(t, msg1.ID)

	msg2 := &bot.Message{
		ConversationID: conv.ID,
		SenderType:     "bot",
		Content:        "Halo Budi, silakan ketik nomor pelanggan Anda.",
		CreatedAt:      time.Now().UTC().Add(time.Second),
	}
	err = repo.CreateMessage(ctx, msg2)
	require.NoError(t, err)

	// 4. FindMessagesByConversationID
	msgs, err := repo.FindMessagesByConversationID(ctx, conv.ID)
	require.NoError(t, err)
	assert.Len(t, msgs, 2)

	// 5. FindRecentMessages
	recent, err := repo.FindRecentMessages(ctx, conv.ID, 1)
	require.NoError(t, err)
	assert.Len(t, recent, 1)
	assert.Equal(t, "Halo Budi, silakan ketik nomor pelanggan Anda.", recent[0].Content)

	// 6. FindConversationByIDWithMessages
	withMsgs, err := repo.FindConversationByIDWithMessages(ctx, conv.ID)
	require.NoError(t, err)
	assert.Len(t, withMsgs.Messages, 2)

	// 7. FindConversationsBySessionID & FindConversationsByStatus
	bySession, err := repo.FindConversationsBySessionID(ctx, 1)
	require.NoError(t, err)
	assert.Len(t, bySession, 1)

	byStatus, err := repo.FindConversationsByStatus(ctx, bot.StatusBot)
	require.NoError(t, err)
	assert.Len(t, byStatus, 1)

	allConvs, err := repo.FindAllConversations(ctx)
	require.NoError(t, err)
	assert.Len(t, allConvs, 1)

	// 8. FindActiveConversationByCustomer
	active, err := repo.FindActiveConversationByCustomer(ctx, 1, "628987654321")
	require.NoError(t, err)
	assert.Equal(t, conv.ID, active.ID)

	// 9. Update Conversation
	conv.Status = bot.StatusDone
	err = repo.UpdateConversation(ctx, conv)
	require.NoError(t, err)

	updatedConv, err := repo.FindConversationByID(ctx, conv.ID)
	require.NoError(t, err)
	assert.Equal(t, bot.StatusDone, updatedConv.Status)
}

