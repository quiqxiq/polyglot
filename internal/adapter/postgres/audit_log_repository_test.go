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
	"github.com/quixiq/polyglot/internal/domain/audit"
)

func setupAuditTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(&model.AuditLogModel{})
	require.NoError(t, err)

	return db
}

func TestAuditLogRepository_Write(t *testing.T) {
	db := setupAuditTestDB(t)
	repo := postgres.NewAuditLogRepository(db)
	ctx := context.Background()

	entry := audit.AuditLog{
		TenantID:    "tenant-default",
		ActorType:   audit.ActorUser,
		ActorID:     "user-001",
		Action:      "REBOOT_DEVICE",
		EntityType:  "device",
		EntityID:    "router-core-1",
		Description: "Scheduled reboot for maintenance",
		IPAddress:   "192.168.1.100",
		CreatedAt:   time.Now().UTC(),
	}

	err := repo.Write(ctx, entry)
	require.NoError(t, err)

	var count int64
	err = db.Model(&model.AuditLogModel{}).Count(&count).Error
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)
}
