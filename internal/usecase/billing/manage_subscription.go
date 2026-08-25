package billing

import (
	"context"
	"fmt"

	domainPlan "github.com/quixiq/polyglot/internal/domain/plan"
	domainSub "github.com/quixiq/polyglot/internal/domain/subscription"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/pkg/logger"
)

// SubscriptionProvisioner is the narrow seam to the network orchestrator,
// implemented by *network.SubscriptionProvisioner. Declared locally so this
// package never imports vendor or driver knowledge.
type SubscriptionProvisioner interface {
	Provision(ctx context.Context, sub domainSub.Subscription, password string) (domainSub.Subscription, error)
	Deprovision(ctx context.Context, subscriptionID string, removeAccount bool) error
	Isolate(ctx context.Context, subscriptionID, reason string) (domainSub.Subscription, error)
	Unisolate(ctx context.Context, subscriptionID string) (domainSub.Subscription, error)
}

type SubscriptionUseCase struct {
	repo        port.SubscriptionRepository
	plans       planLookupForSub
	secrets     port.SecretVault
	provisioner SubscriptionProvisioner
}

type planLookupForSub interface {
	FindByID(ctx context.Context, id string) (domainPlan.Plan, error)
}

func NewSubscriptionUseCase(repo port.SubscriptionRepository) *SubscriptionUseCase {
	return &SubscriptionUseCase{repo: repo}
}

// SetDeps wires the optional collaborators needed by provisioning-aware
// operations (create-with-provision, isolate, delete). Called once at app
// composition.
func (u *SubscriptionUseCase) SetDeps(plans planLookupForSub, secrets port.SecretVault, p SubscriptionProvisioner) {
	u.plans = plans
	u.secrets = secrets
	u.provisioner = p
}

func (u *SubscriptionUseCase) ListSubscriptions(ctx context.Context, customerID string) ([]domainSub.Subscription, error) {
	if u.repo == nil {
		return nil, fmt.Errorf("subscription repository unavailable")
	}
	if customerID != "" {
		return u.repo.FindByCustomerID(ctx, customerID)
	}
	return u.repo.FindAll(ctx)
}

func (u *SubscriptionUseCase) GetSubscription(ctx context.Context, id string) (domainSub.Subscription, error) {
	if u.repo == nil {
		return domainSub.Subscription{}, fmt.Errorf("subscription repository unavailable")
	}
	if id == "" {
		return domainSub.Subscription{}, domainSub.ErrNotFound
	}
	return u.repo.FindByID(ctx, id)
}

// Create persists a PENDING_PROVISION subscription; when device_id and
// remote_username are present AND a provisioner is wired, the account is
// created on the router immediately (mapping saved, status ACTIVE).
func (u *SubscriptionUseCase) CreateSubscription(ctx context.Context, sub domainSub.Subscription, password string) (domainSub.Subscription, error) {
	if u.repo == nil {
		return domainSub.Subscription{}, fmt.Errorf("subscription repository unavailable")
	}
	if err := sub.Validate(); err != nil {
		return domainSub.Subscription{}, err
	}
	planRow, err := u.plans.FindByID(ctx, sub.PlanID)
	if err != nil {
		return domainSub.Subscription{}, fmt.Errorf("load plan: %w", err)
	}
	sub.ServiceType = planRow.ServiceType
	if sub.Status == "" || sub.Status == domainSub.StatusActive {
		sub.Status = domainSub.StatusPendingProvision
	}
	if err := u.repo.Save(ctx, sub); err != nil {
		return domainSub.Subscription{}, err
	}
	if sub.DeviceID != "" && sub.RemoteUsername != "" && u.provisioner != nil {
		return u.provisioner.Provision(ctx, sub, password)
	}
	return sub, nil
}

// UpdateSubscription mutates business/mapping fields. When changeProfile is
// requested the router-side profile is re-derived by un-isolating logic —
// here we simply require ACTIVE status and let the next isolate/unisolate
// cycle refresh the device state.
func (u *SubscriptionUseCase) UpdateSubscription(ctx context.Context, current domainSub.Subscription, planID, username, notes string, billingDay int) (domainSub.Subscription, error) {
	if planID != "" && planID != current.PlanID {
		current.PlanID = planID
	}
	if username != "" {
		current.RemoteUsername = username
	}
	if notes != "" {
		current.Notes = notes
	}
	if billingDay > 0 {
		current.BillingDay = billingDay
	}
	if err := current.Validate(); err != nil {
		return current, err
	}
	if err := u.repo.Save(ctx, current); err != nil {
		return current, fmt.Errorf("update subscription: %w", err)
	}
	logger.WithComponent("ManageSubscriptions").WithField("subscription_id", current.ID).Info("subscription updated")
	return current, nil
}

// DeleteSubscription terminates; deprovision also removes the router-side
// account so no orphan credentials remain on the device.
func (u *SubscriptionUseCase) DeleteSubscription(ctx context.Context, id string, deprovision bool) error {
	if u.repo == nil {
		return fmt.Errorf("subscription repository unavailable")
	}
	if _, err := u.repo.FindByID(ctx, id); err != nil {
		return err
	}
	if deprovision && u.provisioner != nil {
		if err := u.provisioner.Deprovision(ctx, id, true); err != nil {
			return err
		}
		return nil
	}
	return u.repo.UpdateStatus(ctx, id, domainSub.StatusTerminated)
}

func (u *SubscriptionUseCase) CancelSubscription(ctx context.Context, id string) (domainSub.Subscription, error) {
	if u.repo == nil {
		return domainSub.Subscription{}, fmt.Errorf("subscription repository unavailable")
	}
	sub, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return domainSub.Subscription{}, err
	}
	sub.Status = domainSub.StatusTerminated
	if err := u.repo.Save(ctx, sub); err != nil {
		return domainSub.Subscription{}, err
	}
	return sub, nil
}

// Isolate delegates to the provisioner (portal redirect for PPPoE).
func (u *SubscriptionUseCase) IsolateSubscription(ctx context.Context, id, reason string) (domainSub.Subscription, error) {
	if u.provisioner == nil {
		return domainSub.Subscription{}, fmt.Errorf("provisioner not wired")
	}
	return u.provisioner.Isolate(ctx, id, reason)
}

// Unisolate restores normal service.
func (u *SubscriptionUseCase) UnisolateSubscription(ctx context.Context, id string) (domainSub.Subscription, error) {
	if u.provisioner == nil {
		return domainSub.Subscription{}, fmt.Errorf("provisioner not wired")
	}
	return u.provisioner.Unisolate(ctx, id)
}
