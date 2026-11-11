package postgres_test

import (
	"context"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/quixiq/polyglot/internal/adapter/postgres"
	"github.com/quixiq/polyglot/internal/adapter/postgres/model"
	"github.com/quixiq/polyglot/internal/domain/whatsapp"
)

func setupWASessionTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(&model.WASessionModel{}, &model.WAChatModel{}, &model.WAMessageModel{})
	require.NoError(t, err)

	return db
}

func TestWASessionRepository_CRUD(t *testing.T) {
	db := setupWASessionTestDB(t)
	repo := postgres.NewWASessionRepository(db)
	ctx := context.Background()

	// 1. Create Session
	session := &whatsapp.WASession{
		JID:        "628123456789@s.whatsapp.net",
		DeviceName: "Gateway 1",
		Status:     whatsapp.StatusOnline,
	}
	err := repo.CreateSession(ctx, session)
	require.NoError(t, err)
	assert.NotZero(t, session.ID)

	// 2. FindSessionByID
	found, err := repo.FindSessionByID(ctx, session.ID)
	require.NoError(t, err)
	assert.Equal(t, "628123456789@s.whatsapp.net", found.JID)
	assert.Equal(t, whatsapp.StatusOnline, found.Status)

	// 3. FindAllSessions
	all, err := repo.FindAllSessions(ctx)
	require.NoError(t, err)
	assert.Len(t, all, 1)

	// 4. UpdateSession
	session.Status = whatsapp.StatusOffline
	err = repo.UpdateSession(ctx, session)
	require.NoError(t, err)

	afterUp, err := repo.FindSessionByID(ctx, session.ID)
	require.NoError(t, err)
	assert.Equal(t, whatsapp.StatusOffline, afterUp.Status)

	// 5. DeleteSession
	err = repo.DeleteSession(ctx, session.ID)
	require.NoError(t, err)

	_, err = repo.FindSessionByID(ctx, session.ID)
	assert.ErrorIs(t, err, postgres.ErrNotFound)
}

func TestWASessionRepository_UpdateDoesNotResurrectDeletedSession(t *testing.T) {
	db := setupWASessionTestDB(t)
	repo := postgres.NewWASessionRepository(db)
	ctx := context.Background()

	session := &whatsapp.WASession{DeviceName: "Gateway 3", Status: whatsapp.StatusOnline}
	require.NoError(t, repo.CreateSession(ctx, session))

	loaded, err := repo.FindSessionByID(ctx, session.ID)
	require.NoError(t, err)
	require.NoError(t, repo.DeleteSession(ctx, session.ID))

	loaded.Status = whatsapp.StatusNeedsRescan
	err = repo.UpdateSession(ctx, loaded)
	assert.ErrorIs(t, err, postgres.ErrNotFound)

	var count int64
	require.NoError(t, db.Model(&model.WASessionModel{}).Where("id = ?", session.ID).Count(&count).Error)
	assert.Zero(t, count)
}

func TestWASessionRepository_PurgeSession(t *testing.T) {
	db := setupWASessionTestDB(t)
	repo := postgres.NewWASessionRepository(db)
	ctx := context.Background()

	session := &whatsapp.WASession{DeviceName: "Gateway 2", Status: whatsapp.StatusOnline}
	require.NoError(t, repo.CreateSession(ctx, session))

	require.NoError(t, db.Create(&model.WAChatModel{
		SessionID: session.ID,
		ChatJID:   "628123456789@s.whatsapp.net",
	}).Error)
	require.NoError(t, db.Create(&model.WAMessageModel{
		SessionID:   session.ID,
		ChatJID:     "628123456789@s.whatsapp.net",
		WAMessageID: "msg-1",
	}).Error)

	require.NoError(t, repo.PurgeSession(ctx, session.ID))

	_, err := repo.FindSessionByID(ctx, session.ID)
	assert.ErrorIs(t, err, postgres.ErrNotFound)

	var chats, messages int64
	require.NoError(t, db.Model(&model.WAChatModel{}).Where("session_id = ?", session.ID).Count(&chats).Error)
	require.NoError(t, db.Model(&model.WAMessageModel{}).Where("session_id = ?", session.ID).Count(&messages).Error)
	assert.Zero(t, chats)
	assert.Zero(t, messages)
}
