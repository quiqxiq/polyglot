package billing

import (
	"context"
	"fmt"

	domainSub "github.com/quixiq/polyglot/internal/domain/subscription"
	"github.com/quixiq/polyglot/internal/port"
)

type SubscriptionUsecase struct {
	repo port.SubscriptionRepository
}

func NewSubscriptionUsecase(repo port.SubscriptionRepository) *SubscriptionUsecase {
	return &SubscriptionUsecase{repo: repo}
}

func (u *SubscriptionUsecase) ListSubscriptions(ctx context.Context, customerID string) ([]domainSub.Subscription, error) {
	if u.repo == nil {
		return nil, fmt.Errorf("subscription repository unavailable")
	}
	if customerID != "" {
		return u.repo.FindByCustomerID(ctx, customerID)
	}
	return u.repo.FindAll(ctx)
}

func (u *SubscriptionUsecase) GetSubscription(ctx context.Context, id string) (domainSub.Subscription, error) {
	if u.repo == nil {
		return domainSub.Subscription{}, fmt.Errorf("subscription repository unavailable")
	}
	return u.repo.FindByID(ctx, id)
}

func (u *SubscriptionUsecase) CreateSubscription(ctx context.Context, sub domainSub.Subscription) (domainSub.Subscription, error) {
	if u.repo == nil {
		return domainSub.Subscription{}, fmt.Errorf("subscription repository unavailable")
	}
	if sub.CustomerID == "" || sub.PlanID == "" {
		return domainSub.Subscription{}, fmt.Errorf("customer_id and plan_id are required")
	}
	if sub.Status == "" {
		sub.Status = domainSub.StatusActive
	}
	if err := u.repo.Save(ctx, sub); err != nil {
		return domainSub.Subscription{}, err
	}
	return sub, nil
}

func (u *SubscriptionUsecase) CancelSubscription(ctx context.Context, id string) (domainSub.Subscription, error) {
	if u.repo == nil {
		return domainSub.Subscription{}, fmt.Errorf("subscription repository unavailable")
	}
	sub, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return domainSub.Subscription{}, err
	}
	sub.Status = domainSub.StatusCancelled
	if err := u.repo.Save(ctx, sub); err != nil {
		return domainSub.Subscription{}, err
	}
	return sub, nil
}
