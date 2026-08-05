package postgres

import (
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/quixiq/polyglot/internal/adapter/postgres/models"
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
		&models.WASessionModel{},
		&models.LLMConfigModel{},
		&models.ConversationModel{},
		&models.MessageModel{},
		&models.KnowledgeEntryModel{},
		&models.TechnicianModel{},
	); err != nil {
		return nil, fmt.Errorf("failed to auto-migrate: %w", err)
	}

	log.Println("[Postgres Adapter] Database connected and migrated successfully")
	return &Store{db: db}, nil
}

// DB returns the underlying GORM DB instance.
func (s *Store) DB() *gorm.DB {
	return s.db
}
