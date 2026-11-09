package postgres

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/quixiq/polyglot/internal/adapter/postgres/model"
	"github.com/quixiq/polyglot/internal/domain/billing"
	domainTenant "github.com/quixiq/polyglot/internal/domain/tenant"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/pkg/idgen"
)

// PaymentMethodRepository implements port.PaymentMethodRepository backed by PostgreSQL.
type PaymentMethodRepository struct {
	db *gorm.DB
}

var _ port.PaymentMethodRepository = (*PaymentMethodRepository)(nil)

// NewPaymentMethodRepository creates a new port.PaymentMethodRepository.
func NewPaymentMethodRepository(db *gorm.DB) *PaymentMethodRepository {
	return &PaymentMethodRepository{db: db}
}

// Save inserts or updates a payment method.
func (r *PaymentMethodRepository) Save(ctx context.Context, m billing.PaymentMethod) error {
	if m.ID == "" {
		m.ID = idgen.New("pm")
	}
	if m.TenantID == "" {
		m.TenantID = domainTenant.DefaultID
	}
	now := time.Now().UTC()
	if m.CreatedAt.IsZero() {
		m.CreatedAt = now
	}
	m.UpdatedAt = now
	mdl := model.PaymentMethodModelFromDomain(m)
	return r.db.WithContext(ctx).Save(mdl).Error
}

// FindByID retrieves a payment method by ID.
func (r *PaymentMethodRepository) FindByID(ctx context.Context, id string) (billing.PaymentMethod, error) {
	var m model.PaymentMethodModel
	if err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return billing.PaymentMethod{}, billing.ErrNotFound
		}
		return billing.PaymentMethod{}, err
	}
	return m.ToDomain(), nil
}

// List returns payment methods for a tenant, sorted by sort_order ASC, name ASC.
func (r *PaymentMethodRepository) List(ctx context.Context, tenantID string, activeOnly bool) ([]billing.PaymentMethod, error) {
	if tenantID == "" {
		tenantID = domainTenant.DefaultID
	}
	q := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID)
	if activeOnly {
		q = q.Where("is_active = true")
	}
	var list []model.PaymentMethodModel
	if err := q.Order("sort_order ASC, name ASC").Find(&list).Error; err != nil {
		return nil, err
	}
	out := make([]billing.PaymentMethod, len(list))
	for i, m := range list {
		out[i] = m.ToDomain()
	}
	return out, nil
}

// SetActive toggles the is_active status of a payment method.
func (r *PaymentMethodRepository) SetActive(ctx context.Context, id string, active bool) error {
	res := r.db.WithContext(ctx).Model(&model.PaymentMethodModel{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"is_active":  active,
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
