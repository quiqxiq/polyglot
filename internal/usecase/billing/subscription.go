package billing

import (
	"context"
	"fmt"

	domainBilling "github.com/quixiq/polyglot/internal/domain/billing"
	domainCustomer "github.com/quixiq/polyglot/internal/domain/customer"
	domainDevice "github.com/quixiq/polyglot/internal/domain/device"
	domainPlan "github.com/quixiq/polyglot/internal/domain/plan"
	domainSub "github.com/quixiq/polyglot/internal/domain/subscription"
	"github.com/quixiq/polyglot/internal/port"
)

// SubscriptionDetail menggabungkan Subscription dengan metadata Plan, Customer, dan Device
// untuk denormalisasi respon tampilan (Zero Waterfall di Frontend).
type SubscriptionDetail struct {
	Subscription domainSub.Subscription
	Plan         *domainPlan.ServicePlan
	Customer     *domainCustomer.Customer
	Device       *domainDevice.Device
}

// SubscriptionUseCase adalah usecase read/crud yang dipakai
// handler ConnectRPC billing (list/get/create/cancel).
type SubscriptionUseCase struct {
	repo      port.SubscriptionRepository
	plans     port.ServicePlanRepository
	customers port.CustomerRepository
	devices   port.DeviceRepository
}

// NewSubscriptionUseCase membangun usecase langganan dengan dependensi repositori terkait.
func NewSubscriptionUseCase(
	repo port.SubscriptionRepository,
	plans port.ServicePlanRepository,
	customers port.CustomerRepository,
	devices port.DeviceRepository,
) *SubscriptionUseCase {
	return &SubscriptionUseCase{
		repo:      repo,
		plans:     plans,
		customers: customers,
		devices:   devices,
	}
}

// Enrich memperkaya model domain Subscription dengan metadata Plan, Customer, dan Device.
func (u *SubscriptionUseCase) Enrich(ctx context.Context, sub domainSub.Subscription) SubscriptionDetail {
	detail := SubscriptionDetail{Subscription: sub}
	if u.plans != nil && sub.PlanID != "" {
		if pl, err := u.plans.FindByID(ctx, sub.PlanID); err == nil {
			detail.Plan = &pl
		}
	}
	if u.customers != nil && sub.CustomerID != "" {
		if cust, err := u.customers.FindByID(ctx, sub.CustomerID); err == nil {
			detail.Customer = &cust
		}
	}
	if u.devices != nil && sub.DeviceID != nil && *sub.DeviceID != "" {
		if dev, err := u.devices.FindByID(ctx, *sub.DeviceID); err == nil {
			detail.Device = &dev
		}
	}
	return detail
}

// ListSubscriptions mengembalikan daftar langganan yang diperkaya metadata plan/customer/device.
func (u *SubscriptionUseCase) ListSubscriptions(ctx context.Context, customerID string) ([]SubscriptionDetail, error) {
	if u.repo == nil {
		return nil, domainBilling.ErrRepositoryUnavailable
	}
	var subs []domainSub.Subscription
	var err error
	if customerID != "" {
		subs, err = u.repo.FindByCustomerID(ctx, customerID)
	} else {
		subs, err = u.repo.FindAll(ctx)
	}
	if err != nil {
		return nil, fmt.Errorf("find subscriptions: %w", err)
	}

	out := make([]SubscriptionDetail, len(subs))
	for i, sub := range subs {
		out[i] = u.Enrich(ctx, sub)
	}
	return out, nil
}

// GetSubscription mengembalikan detail langganan berdasarkan ID yang diperkaya metadata.
func (u *SubscriptionUseCase) GetSubscription(ctx context.Context, id string) (SubscriptionDetail, error) {
	if u.repo == nil {
		return SubscriptionDetail{}, domainBilling.ErrRepositoryUnavailable
	}
	sub, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return SubscriptionDetail{}, fmt.Errorf("find subscription %s: %w", id, err)
	}
	return u.Enrich(ctx, sub), nil
}

// CreateSubscription membuat langganan baru.
func (u *SubscriptionUseCase) CreateSubscription(ctx context.Context, sub domainSub.Subscription) (domainSub.Subscription, error) {
	if u.repo == nil {
		return domainSub.Subscription{}, domainBilling.ErrRepositoryUnavailable
	}
	if sub.CustomerID == "" || sub.PlanID == "" {
		return domainSub.Subscription{}, domainBilling.ErrInvalidInput
	}
	if sub.Status == "" {
		sub.Status = domainSub.StatusActive
	}
	if err := u.repo.Save(ctx, sub); err != nil {
		return domainSub.Subscription{}, fmt.Errorf("save subscription: %w", err)
	}
	return sub, nil
}

// CancelSubscription membatalkan langganan.
func (u *SubscriptionUseCase) CancelSubscription(ctx context.Context, id string) (domainSub.Subscription, error) {
	if u.repo == nil {
		return domainSub.Subscription{}, domainBilling.ErrRepositoryUnavailable
	}
	sub, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return domainSub.Subscription{}, fmt.Errorf("find subscription %s: %w", id, err)
	}
	sub.Status = domainSub.StatusCancelled
	if err := u.repo.Save(ctx, sub); err != nil {
		return domainSub.Subscription{}, fmt.Errorf("cancel subscription: %w", err)
	}
	return sub, nil
}
