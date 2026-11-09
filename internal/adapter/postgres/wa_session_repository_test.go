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

	err = db.AutoMigrate(&model.WASessionModel{})
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
