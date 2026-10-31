package postgres

import (
	"context"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/quixiq/polyglot/internal/adapter/postgres/model"
	"github.com/quixiq/polyglot/internal/domain/routerjob"
	"github.com/quixiq/polyglot/internal/port"
)

// RouterJobRepository adalah antrean job router berbasis GORM/Postgres.
type RouterJobRepository struct {
	db *gorm.DB
}

var _ port.RouterJobRepository = (*RouterJobRepository)(nil)

// NewRouterJobRepository returns a port.RouterJobRepository backed by GORM.
func NewRouterJobRepository(db *gorm.DB) *RouterJobRepository {
	return &RouterJobRepository{db: db}
}

// Enqueue menambah job secara idempoten berdasarkan idempotency_key.
func (r *RouterJobRepository) Enqueue(ctx context.Context, job routerjob.Job) (routerjob.Job, error) {
	var existing model.RouterJobModel
	err := r.db.WithContext(ctx).Where("idempotency_key = ?", job.IdempotencyKey).First(&existing).Error
	if err == nil {
		return existing.ToDomain(), nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return routerjob.Job{}, err
	}
	m := model.RouterJobModelFromDomain(job)
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		// Race: job dengan key sama dibuat bersamaan → kembalikan yang ada.
		if strings.Contains(strings.ToLower(err.Error()), "unique") {
			var raced model.RouterJobModel
			if ferr := r.db.WithContext(ctx).Where("idempotency_key = ?", job.IdempotencyKey).First(&raced).Error; ferr == nil {
				return raced.ToDomain(), nil
			}
		}
		return routerjob.Job{}, err
	}
	return m.ToDomain(), nil
}

// ClaimDue mengambil batch job siap jalan (atomik; FOR UPDATE SKIP LOCKED di
// Postgres agar beberapa worker tidak mengambil job yang sama).
func (r *RouterJobRepository) ClaimDue(ctx context.Context, limit int, now time.Time) ([]routerjob.Job, error) {
	if limit <= 0 {
		limit = 20
	}
	var out []routerjob.Job
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var rows []model.RouterJobModel
		q := tx.Where("status IN ? AND next_run_at <= ?",
			[]string{string(routerjob.StatusQueued), string(routerjob.StatusRetryWait)}, now).
			Order("next_run_at asc").Limit(limit)
		if r.db.Dialector.Name() == "postgres" {
			q = q.Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"})
		}
		if err := q.Find(&rows).Error; err != nil {
			return err
		}
		for i := range rows {
			rows[i].Status = string(routerjob.StatusRunning)
			rows[i].Attempts++
			rows[i].UpdatedAt = now
			if err := tx.Save(&rows[i]).Error; err != nil {
				return err
			}
			out = append(out, rows[i].ToDomain())
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

// MarkSucceeded menandai job sukses.
func (r *RouterJobRepository) MarkSucceeded(ctx context.Context, id string) error {
	return r.update(ctx, id, map[string]any{
		"status":     string(routerjob.StatusSucceeded),
		"last_error": "",
		"updated_at": time.Now(),
	})
}

// MarkRetry menandai job menunggu percobaan berikutnya.
func (r *RouterJobRepository) MarkRetry(ctx context.Context, id string, nextRunAt time.Time, lastErr string) error {
	return r.update(ctx, id, map[string]any{
		"status":      string(routerjob.StatusRetryWait),
		"next_run_at": nextRunAt,
		"last_error":  lastErr,
		"updated_at":  time.Now(),
	})
}

// MarkDead menandai job gagal permanen.
func (r *RouterJobRepository) MarkDead(ctx context.Context, id string, lastErr string) error {
	return r.update(ctx, id, map[string]any{
		"status":     string(routerjob.StatusDead),
		"last_error": lastErr,
		"updated_at": time.Now(),
	})
}

// Requeue mengembalikan job DEAD/FAILED ke QUEUED untuk retry manual.
func (r *RouterJobRepository) Requeue(ctx context.Context, id string) error {
	return r.update(ctx, id, map[string]any{
		"status":      string(routerjob.StatusQueued),
		"attempts":    0,
		"last_error":  "",
		"next_run_at": time.Now(),
		"updated_at":  time.Now(),
	})
}

func (r *RouterJobRepository) update(ctx context.Context, id string, updates map[string]any) error {
	res := r.db.WithContext(ctx).Model(&model.RouterJobModel{}).Where("id = ?", id).Updates(updates)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// FindByID mencari job berdasarkan id.
func (r *RouterJobRepository) FindByID(ctx context.Context, id string) (routerjob.Job, error) {
	var m model.RouterJobModel
	if err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return routerjob.Job{}, ErrNotFound
		}
		return routerjob.Job{}, err
	}
	return m.ToDomain(), nil
}

// List mengembalikan job terbaru, opsional difilter status.
func (r *RouterJobRepository) List(ctx context.Context, f port.RouterJobFilter) ([]routerjob.Job, error) {
	q := r.db.WithContext(ctx).Model(&model.RouterJobModel{})
	if f.Status != "" {
		q = q.Where("status = ?", f.Status)
	}
	q = q.Order("created_at desc")
	if f.Limit > 0 {
		q = q.Limit(f.Limit)
	}
	var rows []model.RouterJobModel
	if err := q.Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]routerjob.Job, len(rows))
	for i := range rows {
		out[i] = rows[i].ToDomain()
	}
	return out, nil
}
