package postgres

import (
	"context"
	"hash/fnv"

	"gorm.io/gorm"

	"github.com/quixiq/polyglot/internal/port"
)

// AdvisoryLocker mengimplementasikan port.JobLocker dengan
// pg_try_advisory_lock pada koneksi database yang dipegang selama job
// berjalan. Dialect non-Postgres (sqlite test) selalu sukses (F6-1).
type AdvisoryLocker struct {
	db *gorm.DB
}

var _ port.JobLocker = (*AdvisoryLocker)(nil)

// NewAdvisoryLocker constructs a JobLocker backed by GORM.
func NewAdvisoryLocker(db *gorm.DB) *AdvisoryLocker {
	return &AdvisoryLocker{db: db}
}

// TryLock implements port.JobLocker.
func (l *AdvisoryLocker) TryLock(ctx context.Context, key string) (func(), bool, error) {
	noop := func() {}
	if l.db.Name() != "postgres" {
		return noop, true, nil
	}
	sqlDB, err := l.db.DB()
	if err != nil {
		return noop, false, err
	}
	conn, err := sqlDB.Conn(ctx)
	if err != nil {
		return noop, false, err
	}
	var got bool
	if err := conn.QueryRowContext(ctx, "SELECT pg_try_advisory_lock($1)", advisoryKey(key)).Scan(&got); err != nil {
		_ = conn.Close()
		return noop, false, err
	}
	if !got {
		_ = conn.Close()
		return noop, false, nil
	}
	unlock := func() {
		_, _ = conn.ExecContext(context.Background(), "SELECT pg_advisory_unlock($1)", advisoryKey(key))
		_ = conn.Close()
	}
	return unlock, true, nil
}

// advisoryKey menurunkan key int64 stabil dari nama job.
func advisoryKey(name string) int64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(name))
	return int64(h.Sum64())
}
