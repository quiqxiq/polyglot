package postgres

import (
	"context"
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB wraps a GORM connection to PostgreSQL. It exposes the raw *gorm.DB
// for repository construction and lifecycle helpers for Ping/Close.
type DB struct {
	inner *gorm.DB
}

// NewDB opens a GORM connection to PostgreSQL using the supplied DSN.
// It configures sensible connection pool defaults and enables a logger
// that is silent by default.
func NewDB(ctx context.Context, dsn string) (*DB, error) {
	gormDB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("postgres: open connection: %w", err)
	}

	sqlDB, err := gormDB.DB()
	if err != nil {
		return nil, fmt.Errorf("postgres: get underlying sql db: %w", err)
	}

	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)

	return &DB{inner: gormDB}, nil
}

// GORM returns the underlying *gorm.DB for repository construction.
func (db *DB) GORM() *gorm.DB {
	return db.inner
}

// Ping verifies the database connection is alive.
func (db *DB) Ping(ctx context.Context) error {
	sqlDB, err := db.inner.DB()
	if err != nil {
		return fmt.Errorf("postgres: get underlying sql db: %w", err)
	}
	return sqlDB.PingContext(ctx)
}

// Close closes the underlying connection pool.
func (db *DB) Close() error {
	sqlDB, err := db.inner.DB()
	if err != nil {
		return fmt.Errorf("postgres: get underlying sql db: %w", err)
	}
	return sqlDB.Close()
}
