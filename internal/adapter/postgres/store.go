package postgres

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/quixiq/polyglot/pkg/logger"

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
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get database connection pool: %w", err)
	}
	sqlDB.SetMaxOpenConns(50)
	sqlDB.SetMaxIdleConns(25)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)
	sqlDB.SetConnMaxIdleTime(5 * time.Minute)
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}

	appEnv := strings.ToLower(os.Getenv("APP_ENV"))
	if appEnv == "development" || appEnv == "" || os.Getenv("AUTO_MIGRATE") == "true" {
		if err := db.AutoMigrate(
			&model.UserModel{},
			&model.UserDeviceModel{},
			&model.WASessionModel{},
			&model.LLMConfigModel{},
			&model.ConversationModel{},
			&model.MessageModel{},
			&model.SkillMetadataModel{},
			&model.GlobalPromptModel{},
			&model.DeviceModel{},
			&model.DevicePingMetricModel{},
			&model.CredentialModel{},
			&model.WAChatModel{},
			&model.WAMessageModel{},
			&model.CustomerModel{},
			&model.PortalSessionModel{},
			&model.PortalOTPModel{},
			&model.SubscriptionModel{},
			&model.InvoiceModel{},
			&model.InvoiceItemModel{},
			&model.PaymentMethodModel{},
			&model.PaymentModel{},
			&model.GatewayTransactionModel{},
			&model.ServicePlanModel{},
			&model.RegistrationModel{},
			&model.RegistrationEventModel{},
			&model.PaymentSubmissionModel{},
			&model.RouterJobModel{},
			&model.CashAccountModel{},
			&model.CashCategoryModel{},
			&model.CashTransactionModel{},
			&model.NotificationTemplateModel{},
			&model.WANotificationModel{},
			&model.DailySnapshotModel{},
			&model.AuditLogModel{},
			&model.SystemSettingModel{},
			// Fase A PLAN-ACS-OLT-HIOSO.md: inventori ONT/OLT + cache ACS.
			&model.ONTModel{},
			&model.ONTEventModel{},
			&model.ONTMetricSampleModel{},
			&model.ACSDeviceModel{},
			&model.ACSPresetModel{},
			// W0 PLAN-BILLING-OPS-V3: hutang, batch mass payment, gateway settlement.
			&model.DebtModel{},
			&model.DebtPaymentModel{},
			&model.BulkPaymentBatchModel{},
			&model.BulkPaymentItemModel{},
			&model.GatewaySettlementModel{},
			// M1 PLAN-AI-ROUTER-OPS v2: restore point, approval, audit, active device.
			&model.RestorePointModel{},
			&model.ArtifactModel{},
			&model.ApprovalModel{},
			&model.CommandLogModel{},
			&model.ActiveDeviceModel{},
		); err != nil {
			return nil, fmt.Errorf("failed to auto-migrate: %w", err)
		}
		logger.WithComponent("PostgresAdapter").Info("Auto-migration executed in development mode")
	} else {
		logger.WithComponent("PostgresAdapter").Info("Skipping AutoMigrate in production; relying on SQL migrations")
	}

	// Memastikan kompatibilitas tipe kolom user_devices.device_id dan inisialisasi TimescaleDB hypertable
	if db.Dialector.Name() == "postgres" {
		_ = db.Exec(`DO $$ BEGIN
			IF EXISTS (
				SELECT 1 FROM information_schema.columns 
				WHERE table_name = 'user_devices' 
				  AND column_name = 'device_id' 
				  AND data_type = 'character varying'
			) THEN
				ALTER TABLE user_devices ALTER COLUMN device_id TYPE uuid USING device_id::uuid;
			END IF;
		END $$;`).Error

		_ = db.Exec(`DO $$ BEGIN
			BEGIN
				CREATE EXTENSION IF NOT EXISTS timescaledb;
			EXCEPTION
				WHEN OTHERS THEN
					NULL;
			END;

			IF EXISTS (SELECT 1 FROM pg_extension WHERE extname = 'timescaledb') THEN
				IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'device_ping_metrics') THEN
					ALTER TABLE device_ping_metrics DROP CONSTRAINT IF EXISTS device_ping_metrics_pkey;
				END IF;
				PERFORM create_hypertable('device_ping_metrics', 'recorded_at', if_not_exists => TRUE);
			END IF;
		END $$;`).Error
	}

	logger.WithComponent("PostgresAdapter").Info("Database connected successfully")
	return &Store{db: db}, nil
}

// DB returns the underlying GORM DB instance.
func (s *Store) DB() *gorm.DB {
	return s.db
}
