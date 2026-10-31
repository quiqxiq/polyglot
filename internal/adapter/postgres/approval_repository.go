package postgres

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/quixiq/polyglot/internal/adapter/postgres/model"
	"github.com/quixiq/polyglot/internal/domain/approval"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/pkg/idgen"
)

// approvalDefaultTTL adalah masa berlaku pengaman bila usecase tidak mengisi
// ExpiresAt; keputusan TTL bisnis tetap milik usecase.
const approvalDefaultTTL = 15 * time.Minute

// ApprovalRepository adalah implementasi port.ApprovalRepository berbasis
// GORM/Postgres untuk approval HITL mutasi router (PLAN-AI-ROUTER-OPS v2).
// Semua perubahan status memakai compare-and-set revision agar dua
// approver/eksekutor tidak saling menimpa.
type ApprovalRepository struct {
	db *gorm.DB
}

var _ port.ApprovalRepository = (*ApprovalRepository)(nil)

// NewApprovalRepository returns a port.ApprovalRepository backed by GORM.
func NewApprovalRepository(db *gorm.DB) *ApprovalRepository {
	return &ApprovalRepository{db: db}
}

// Create menyimpan pengajuan approval baru; ID kosong dibuat otomatis.
func (r *ApprovalRepository) Create(ctx context.Context, req approval.Request) error {
	if req.ID == "" {
		req.ID = idgen.New("appr")
	}
	if req.TenantID == "" {
		req.TenantID = defaultTenantID
	}
	if req.Status == "" {
		req.Status = approval.StatusPending
	}
	if req.Revision <= 0 {
		req.Revision = 1
	}
	if req.CreatedAt.IsZero() {
		req.CreatedAt = time.Now().UTC()
	}
	if req.ExpiresAt.IsZero() {
		req.ExpiresAt = req.CreatedAt.Add(approvalDefaultTTL)
	}
	return r.db.WithContext(ctx).Create(model.ApprovalModelFromDomain(req)).Error
}

// FindByID mengambil satu approval berdasarkan id.
func (r *ApprovalRepository) FindByID(ctx context.Context, id string) (approval.Request, error) {
	var m model.ApprovalModel
	if err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error; err != nil {
		return approval.Request{}, mapApprovalError(err)
	}
	return m.ToDomain(), nil
}

// FindByIdempotencyKey mencari approval berdasarkan kunci idempotensi tenant.
func (r *ApprovalRepository) FindByIdempotencyKey(ctx context.Context, tenantID, key string) (approval.Request, error) {
	if key == "" {
		return approval.Request{}, approval.ErrNotFound
	}
	var m model.ApprovalModel
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND idempotency_key = ?", orDefault(tenantID, defaultTenantID), key).
		First(&m).Error
	if err != nil {
		return approval.Request{}, mapApprovalError(err)
	}
	return m.ToDomain(), nil
}

// List mengembalikan approval sesuai filter beserta total baris (sebelum
// limit/offset), terbaru lebih dulu.
func (r *ApprovalRepository) List(ctx context.Context, f port.ApprovalFilter) ([]approval.Request, int, error) {
	q := r.db.WithContext(ctx).Model(&model.ApprovalModel{}).
		Where("tenant_id = ?", orDefault(f.TenantID, defaultTenantID))
	if f.DeviceID != "" {
		q = q.Where("device_id = ?", f.DeviceID)
	}
	if f.Status != "" {
		q = q.Where("status = ?", f.Status)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	q = q.Order("created_at desc, id desc")
	if f.Limit > 0 {
		q = q.Limit(f.Limit)
	}
	if f.Offset > 0 {
		q = q.Offset(f.Offset)
	}
	var rows []model.ApprovalModel
	if err := q.Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	out := make([]approval.Request, len(rows))
	for i := range rows {
		out[i] = rows[i].ToDomain()
	}
	return out, int(total), nil
}

// UpdateIfRevision menyimpan req hanya bila revision tersimpan sama dengan
// expectedRevision. Return false (tanpa error) pada CAS mismatch.
func (r *ApprovalRepository) UpdateIfRevision(ctx context.Context, req approval.Request, expectedRevision int) (bool, error) {
	if req.ID == "" {
		return false, approval.ErrNotFound
	}
	m := model.ApprovalModelFromDomain(req)
	res := r.db.WithContext(ctx).
		Model(&model.ApprovalModel{}).
		Where("id = ? AND revision = ?", req.ID, expectedRevision).
		Select("*").
		Omit("id", "created_at", "updated_at").
		Updates(m)
	if res.Error != nil {
		return false, res.Error
	}
	if res.RowsAffected > 0 {
		return true, nil
	}
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.ApprovalModel{}).
		Where("id = ?", req.ID).Count(&count).Error; err != nil {
		return false, err
	}
	if count == 0 {
		return false, approval.ErrNotFound
	}
	return false, nil
}

// ClaimForExecution mengklaim APPROVED→EXECUTING secara atomik dengan row
// lock, mengisi execution_nonce bila kosong, dan menaikkan revision.
func (r *ApprovalRepository) ClaimForExecution(ctx context.Context, id string, now time.Time) (approval.Request, bool, error) {
	var out approval.Request
	claimed := false
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		q := tx
		if r.db.Dialector.Name() == "postgres" {
			q = q.Clauses(lockingClause())
		}
		var m model.ApprovalModel
		if err := q.First(&m, "id = ?", id).Error; err != nil {
			return mapApprovalError(err)
		}
		req := m.ToDomain()
		if !req.CanExecute(now) {
			return nil
		}
		nonce := req.ExecutionNonce
		if nonce == "" {
			nonce = idgen.New("exec")
		}
		if err := tx.Model(&model.ApprovalModel{}).Where("id = ?", id).
			Updates(map[string]any{
				"status":          string(approval.StatusExecuting),
				"execution_nonce": nonce,
				"revision":        gorm.Expr("revision + 1"),
			}).Error; err != nil {
			return err
		}
		// Muat ulang agar UpdatedAt/Revision konsisten dengan trigger DB.
		if err := tx.First(&m, "id = ?", id).Error; err != nil {
			return mapApprovalError(err)
		}
		out = m.ToDomain()
		claimed = true
		return nil
	})
	if err != nil {
		return approval.Request{}, false, err
	}
	return out, claimed, nil
}

// ExpirePending menandai PENDING yang melewati expires_at menjadi EXPIRED
// dan mengembalikan jumlah baris yang berubah.
func (r *ApprovalRepository) ExpirePending(ctx context.Context, now time.Time) (int, error) {
	res := r.db.WithContext(ctx).
		Model(&model.ApprovalModel{}).
		Where("status = ? AND expires_at < ?", string(approval.StatusPending), now).
		Updates(map[string]any{
			"status":   string(approval.StatusExpired),
			"revision": gorm.Expr("revision + 1"),
		})
	if res.Error != nil {
		return 0, res.Error
	}
	return int(res.RowsAffected), nil
}

// mapApprovalError menerjemahkan gorm not-found menjadi sentinel domain.
func mapApprovalError(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return approval.ErrNotFound
	}
	return err
}
