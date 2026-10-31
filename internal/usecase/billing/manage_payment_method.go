package billing

import (
	"context"
	"fmt"
	"strings"

	domainBilling "github.com/quixiq/polyglot/internal/domain/billing"
	"github.com/quixiq/polyglot/internal/port"
)

// PaymentMethodUseCase provides CRUD operations for master payment methods.
type PaymentMethodUseCase struct {
	repo port.PaymentMethodRepository
}

// NewPaymentMethodUseCase creates a new PaymentMethodUseCase.
func NewPaymentMethodUseCase(repo port.PaymentMethodRepository) *PaymentMethodUseCase {
	return &PaymentMethodUseCase{repo: repo}
}

// ListPaymentMethods returns payment methods for a tenant.
func (u *PaymentMethodUseCase) ListPaymentMethods(ctx context.Context, tenantID string, activeOnly bool) ([]domainBilling.PaymentMethod, error) {
	if u.repo == nil {
		return nil, domainBilling.ErrRepositoryUnavailable
	}
	return u.repo.List(ctx, tenantID, activeOnly)
}

// SavePaymentMethod creates or updates a payment method.
func (u *PaymentMethodUseCase) SavePaymentMethod(ctx context.Context, m domainBilling.PaymentMethod) (domainBilling.PaymentMethod, error) {
	if u.repo == nil {
		return domainBilling.PaymentMethod{}, domainBilling.ErrRepositoryUnavailable
	}
	if strings.TrimSpace(m.Name) == "" {
		return domainBilling.PaymentMethod{}, fmt.Errorf("%w: name is required", domainBilling.ErrInvalidInput)
	}
	if strings.TrimSpace(m.Type) == "" {
		m.Type = domainBilling.MethodTypeBank
	}
	if err := u.repo.Save(ctx, m); err != nil {
		return domainBilling.PaymentMethod{}, err
	}
	return m, nil
}

// DeletePaymentMethod soft-deactivates a payment method.
func (u *PaymentMethodUseCase) DeletePaymentMethod(ctx context.Context, id string) error {
	if u.repo == nil {
		return domainBilling.ErrRepositoryUnavailable
	}
	return u.repo.SetActive(ctx, id, false)
}
