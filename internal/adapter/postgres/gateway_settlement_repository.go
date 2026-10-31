package postgres

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/quixiq/polyglot/internal/adapter/postgres/model"
	"github.com/quixiq/polyglot/internal/domain/billing"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/pkg/idgen"
)

// GatewaySettlementRepository implements port.GatewaySettlementRepository backed by PostgreSQL.
type GatewaySettlementRepository struct {
	db *gorm.DB
}

var _ port.GatewaySettlementRepository = (*GatewaySettlementRepository)(nil)

// NewGatewaySettlementRepository creates a new port.GatewaySettlementRepository.
func NewGatewaySettlementRepository(db *gorm.DB) *GatewaySettlementRepository {
	return &GatewaySettlementRepository{db: db}
}

// Save stores or updates a gateway settlement record.
func (r *GatewaySettlementRepository) Save(ctx context.Context, s billing.GatewaySettlement) error {
	if s.ID == "" {
		s.ID = idgen.New("gws")
	}
	if s.TenantID == "" {
		s.TenantID = "tenant-default"
	}
	now := time.Now().UTC()
	if s.CreatedAt.IsZero() {
		s.CreatedAt = now
	}
	s.UpdatedAt = now
	mdl := model.GatewaySettlementModelFromDomain(s)
	return r.db.WithContext(ctx).Save(mdl).Error
}

// FindByID retrieves a gateway settlement by ID.
func (r *GatewaySettlementRepository) FindByID(ctx context.Context, id string) (billing.GatewaySettlement, error) {
	var m model.GatewaySettlementModel
	if err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return billing.GatewaySettlement{}, billing.ErrNotFound
		}
		return billing.GatewaySettlement{}, err
	}
	return m.ToDomain(), nil
}

// FindByRef retrieves a settlement by tenant, gateway, and settlement_ref.
func (r *GatewaySettlementRepository) FindByRef(ctx context.Context, tenantID, gateway, ref string) (billing.GatewaySettlement, error) {
	if tenantID == "" {
		tenantID = "tenant-default"
	}
	var m model.GatewaySettlementModel
	if err := r.db.WithContext(ctx).
		First(&m, "tenant_id = ? AND gateway = ? AND settlement_ref = ?", tenantID, gateway, ref).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return billing.GatewaySettlement{}, billing.ErrNotFound
		}
		return billing.GatewaySettlement{}, err
	}
	return m.ToDomain(), nil
}

// List returns filtered gateway settlement records.
func (r *GatewaySettlementRepository) List(ctx context.Context, f port.GatewaySettlementFilter) ([]billing.GatewaySettlement, error) {
	tenantID := f.TenantID
	if tenantID == "" {
		tenantID = "tenant-default"
	}
	q := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID)
	if f.Gateway != "" {
		q = q.Where("gateway = ?", f.Gateway)
	}
	if f.Status != "" {
		q = q.Where("status = ?", f.Status)
	}
	if f.From != nil {
		q = q.Where("period_to >= ?", f.From)
	}
	if f.To != nil {
		q = q.Where("period_from <= ?", f.To)
	}

	limit := f.Limit
	if limit <= 0 {
		limit = 100
	}
	var list []model.GatewaySettlementModel
	if err := q.Order("created_at DESC").Limit(limit).Find(&list).Error; err != nil {
		return nil, err
	}
	out := make([]billing.GatewaySettlement, len(list))
	for i, m := range list {
		out[i] = m.ToDomain()
	}
	return out, nil
}
