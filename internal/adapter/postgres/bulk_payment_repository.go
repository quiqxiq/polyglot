package postgres

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/quixiq/polyglot/internal/adapter/postgres/model"
	"github.com/quixiq/polyglot/internal/domain/billing"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/pkg/idgen"
)

// BulkPaymentRepository implements port.BulkPaymentRepository backed by PostgreSQL.
type BulkPaymentRepository struct {
	db *gorm.DB
}

var _ port.BulkPaymentRepository = (*BulkPaymentRepository)(nil)

// NewBulkPaymentRepository creates a new port.BulkPaymentRepository.
func NewBulkPaymentRepository(db *gorm.DB) *BulkPaymentRepository {
	return &BulkPaymentRepository{db: db}
}

// SaveBatch stores or updates a bulk payment batch.
func (r *BulkPaymentRepository) SaveBatch(ctx context.Context, b billing.BulkPaymentBatch) error {
	if b.ID == "" {
		b.ID = idgen.New("bpb")
	}
	if b.TenantID == "" {
		b.TenantID = "tenant-default"
	}
	now := time.Now().UTC()
	if b.CreatedAt.IsZero() {
		b.CreatedAt = now
	}
	b.UpdatedAt = now
	mdl := model.BulkPaymentBatchModelFromDomain(b)
	return r.db.WithContext(ctx).Save(mdl).Error
}

// FindBatchByID retrieves a batch by ID.
func (r *BulkPaymentRepository) FindBatchByID(ctx context.Context, id string) (billing.BulkPaymentBatch, error) {
	var m model.BulkPaymentBatchModel
	if err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return billing.BulkPaymentBatch{}, billing.ErrNotFound
		}
		return billing.BulkPaymentBatch{}, err
	}
	return m.ToDomain(), nil
}

// ListBatches lists recent batches for a tenant.
func (r *BulkPaymentRepository) ListBatches(ctx context.Context, tenantID string, limit int) ([]billing.BulkPaymentBatch, error) {
	if tenantID == "" {
		tenantID = "tenant-default"
	}
	if limit <= 0 {
		limit = 50
	}
	var list []model.BulkPaymentBatchModel
	if err := r.db.WithContext(ctx).
		Where("tenant_id = ?", tenantID).
		Order("created_at DESC").
		Limit(limit).
		Find(&list).Error; err != nil {
		return nil, err
	}
	out := make([]billing.BulkPaymentBatch, len(list))
	for i, m := range list {
		out[i] = m.ToDomain()
	}
	return out, nil
}

// UpdateBatchStatus modifies status of a batch.
func (r *BulkPaymentRepository) UpdateBatchStatus(ctx context.Context, id, status string) error {
	res := r.db.WithContext(ctx).Model(&model.BulkPaymentBatchModel{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"status":     status,
			"updated_at": time.Now().UTC(),
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return billing.ErrNotFound
	}
	return nil
}

// SaveItems stores item records in batch.
func (r *BulkPaymentRepository) SaveItems(ctx context.Context, items []billing.BulkPaymentItem) error {
	if len(items) == 0 {
		return nil
	}
	now := time.Now().UTC()
	models := make([]model.BulkPaymentItemModel, len(items))
	for i, item := range items {
		if item.ID == "" {
			item.ID = idgen.New("bpi")
		}
		if item.CreatedAt.IsZero() {
			item.CreatedAt = now
		}
		if item.Status == "" {
			item.Status = billing.BulkItemQueued
		}
		models[i] = *model.BulkPaymentItemModelFromDomain(item)
	}
	return r.db.WithContext(ctx).Create(&models).Error
}

// ListItems returns all items for a batch.
func (r *BulkPaymentRepository) ListItems(ctx context.Context, batchID string) ([]billing.BulkPaymentItem, error) {
	var list []model.BulkPaymentItemModel
	if err := r.db.WithContext(ctx).
		Where("batch_id = ?", batchID).
		Order("created_at ASC").
		Find(&list).Error; err != nil {
		return nil, err
	}
	out := make([]billing.BulkPaymentItem, len(list))
	for i, m := range list {
		out[i] = m.ToDomain()
	}
	return out, nil
}

// ClaimNextQueuedItem mengambil satu item QUEUED dan mengubah statusnya menjadi PROCESSING secara atomik.
func (r *BulkPaymentRepository) ClaimNextQueuedItem(ctx context.Context, batchID string) (billing.BulkPaymentItem, bool, error) {
	var claimed billing.BulkPaymentItem
	var found bool

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		q := tx.Model(&model.BulkPaymentItemModel{}).
			Where("batch_id = ? AND status = ?", batchID, billing.BulkItemQueued).
			Order("created_at ASC")

		if r.db.Dialector.Name() == "postgres" {
			q = q.Clauses(clause.Locking{
				Strength: "UPDATE",
				Options:  "SKIP LOCKED",
			})
		}

		var m model.BulkPaymentItemModel
		if err := q.First(&m).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				found = false
				return nil
			}
			return err
		}

		if err := tx.Model(&model.BulkPaymentItemModel{}).
			Where("id = ?", m.ID).
			Update("status", billing.BulkItemProcessing).Error; err != nil {
			return err
		}

		m.Status = billing.BulkItemProcessing
		claimed = m.ToDomain()
		found = true
		return nil
	})

	if err != nil {
		return billing.BulkPaymentItem{}, false, err
	}
	return claimed, found, nil
}

// UpdateItem updates an item's status, error, payment_id, processed_at.
func (r *BulkPaymentRepository) UpdateItem(ctx context.Context, item billing.BulkPaymentItem) error {
	m := model.BulkPaymentItemModelFromDomain(item)
	return r.db.WithContext(ctx).Save(m).Error
}

// CountItems returns counts of items grouped by status.
func (r *BulkPaymentRepository) CountItems(ctx context.Context, batchID string) (queued, processing, succeeded, failed int, err error) {
	type countRow struct {
		Status string
		Count  int
	}
	var rows []countRow
	if err := r.db.WithContext(ctx).Model(&model.BulkPaymentItemModel{}).
		Select("status, COUNT(*) as count").
		Where("batch_id = ?", batchID).
		Group("status").
		Scan(&rows).Error; err != nil {
		return 0, 0, 0, 0, err
	}
	for _, row := range rows {
		switch row.Status {
		case billing.BulkItemQueued:
			queued = row.Count
		case billing.BulkItemProcessing:
			processing = row.Count
		case billing.BulkItemSucceeded:
			succeeded = row.Count
		case billing.BulkItemFailed:
			failed = row.Count
		}
	}
	return queued, processing, succeeded, failed, nil
}
