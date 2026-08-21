package postgres

import (
	"fmt"
	"github.com/quixiq/polyglot/pkg/logger"
	"os"
	"strings"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/quixiq/polyglot/internal/adapter/postgres/model"
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
			&model.UserModel{},
			&model.WASessionModel{},
			&model.LLMConfigModel{},
			&model.ConversationModel{},
			&model.MessageModel{},
			&model.SkillMetadataModel{},
			&model.GlobalPromptModel{},
			&model.DeviceModel{},
			&model.CredentialModel{},
			&model.WAChatModel{},
			&model.WAMessageModel{},
			&model.CustomerModel{},
			&model.SubscriptionModel{},
			&model.InvoiceModel{},
			&model.PlanModel{},
			&model.SystemSettingModel{},
		); err != nil {
			return nil, fmt.Errorf("failed to auto-migrate: %w", err)
		}
		logger.WithComponent("PostgresAdapter").Info("Auto-migration executed in development mode")
	} else {
		logger.WithComponent("PostgresAdapter").Info("Skipping AutoMigrate in production; relying on SQL migrations")
	}

	logger.WithComponent("PostgresAdapter").Info("Database connected successfully")
	return &Store{db: db}, nil
}

// DB returns the underlying GORM DB instance.
func (s *Store) DB() *gorm.DB {
	return s.db
}
