package postgres

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/quixiq/polyglot/internal/adapter/postgres/models"
	"github.com/quixiq/polyglot/pkg/logger"
)

// Store holds the GORM database connection and implements repository interfaces.
type Store struct {
	db *gorm.DB
}

// NewStore opens a PostgreSQL connection and auto-migrates database models.
func NewStore(dsn string) (*Store, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := db.AutoMigrate(
		&models.UserModel{},
		&models.CustomerModel{},
		&models.SubscriptionModel{},
		&models.WASessionModel{},
		&models.LLMConfigModel{},
		&models.ConversationModel{},
		&models.MessageModel{},
		&models.KnowledgeEntryModel{},
		&models.TechnicianModel{},
		&models.DeviceModel{},
		&models.CredentialModel{},
	); err != nil {
		return nil, fmt.Errorf("failed to auto-migrate: %w", err)
	}

	logger.Info("Postgres database connected and migrated successfully")
	return &Store{db: db}, nil
}

// DB returns the underlying GORM DB instance.
func (s *Store) DB() *gorm.DB {
	return s.db
}
