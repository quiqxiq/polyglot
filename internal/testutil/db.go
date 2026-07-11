package testutil

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// NewMemoryDB returns an in-memory SQLite-backed GORM instance for unit
// tests. It is silent and uses a single shared connection so the database
// persists for the lifetime of the returned *gorm.DB.
func NewMemoryDB() (*gorm.DB, error) {
	return gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
}
