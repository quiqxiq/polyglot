package port

import (
	"context"
	"time"

	"github.com/quixiq/polyglot/internal/domain/routerjob"
)

// RouterJobFilter menyaring daftar job router.
type RouterJobFilter struct {
	Status string
	Limit  int
}

// RouterJobRepository adalah antrean job asynchronous per perangkat router
// (planning v2 Fase 4): claim batch, sukses, retry, dead, dan requeue manual.
type RouterJobRepository interface {
	// Enqueue menambah job; bila idempotency_key sudah ada, job lama
	// dikembalikan tanpa error (idempoten).
	Enqueue(ctx context.Context, job routerjob.Job) (routerjob.Job, error)

	// ClaimDue mengambil maksimum limit job yang siap dijalankan
	// (status QUEUED/RETRY_WAIT dan next_run_at <= now), menandainya RUNNING
	// dan menaikkan attempts secara atomik.
	ClaimDue(ctx context.Context, limit int, now time.Time) ([]routerjob.Job, error)

	// MarkSucceeded menandai job sukses.
	MarkSucceeded(ctx context.Context, id string) error
	// MarkRetry menandai job menunggu percobaan berikutnya.
	MarkRetry(ctx context.Context, id string, nextRunAt time.Time, lastErr string) error
	// MarkDead menandai job gagal permanen setelah melewati max attempts.
	MarkDead(ctx context.Context, id string, lastErr string) error
	// Requeue mengembalikan job DEAD/FAILED ke QUEUED untuk retry manual.
	Requeue(ctx context.Context, id string) error

	FindByID(ctx context.Context, id string) (routerjob.Job, error)
	List(ctx context.Context, f RouterJobFilter) ([]routerjob.Job, error)
}
