package postgres

import (
	"fmt"
	"log"
	"os"
	"strings"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/quixiq/polyglot/internal/adapter/postgres/models"
)

// Store holds the GORM database connection and implements repository interfaces.
type Store struct {
	db *gorm.DB
}

// NewStore opens a PostgreSQL connection and optionally auto-migrates database models in development.
func NewStore(dsn string) (*Store, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	appEnv := strings.ToLower(os.Getenv("APP_ENV"))
	if appEnv == "development" || appEnv == "" || os.Getenv("AUTO_MIGRATE") == "true" {
		if err := db.AutoMigrate(
			&models.UserModel{},
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
		log.Println("[Postgres Adapter] Auto-migration executed in development mode")
	} else {
		log.Println("[Postgres Adapter] Skipping AutoMigrate in production; relying on SQL migrations")
	}

	log.Println("[Postgres Adapter] Database connected successfully")
	return &Store{db: db}, nil
}

// DB returns the underlying GORM DB instance.
func (s *Store) DB() *gorm.DB {
	return s.db
}
