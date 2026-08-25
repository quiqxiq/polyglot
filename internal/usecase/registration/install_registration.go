package registration

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/google/uuid"

	domainCustomer "github.com/quixiq/polyglot/internal/domain/customer"
	domainReg "github.com/quixiq/polyglot/internal/domain/registration"
	domainSub "github.com/quixiq/polyglot/internal/domain/subscription"
	"github.com/quixiq/polyglot/pkg/logger"
)

// newID returns a fresh UUID string.
func newID() string { return uuid.NewString() }

// MarkInstalledInput carries the technician's field result.
type MarkInstalledInput struct {
	ID              string
	DeviceID        string
	RemoteUsername  string // opsional; digenerate dari nama bila kosong
	Password        string // opsional; digenerate bila kosong
	TechnicianNotes string
}

// MarkInstalledResult reports the converted entities.
type MarkInstalledResult struct {
	Registration   domainReg.Registration
	CustomerID     string
	SubscriptionID string
}

// Install converts an APPROVED registration into a Customer + Subscription
// and provisions the account on the chosen router. Idempotent under retry:
// a partially-completed run reuses the previously created customer and
// subscription instead of duplicating them, and the provisioner adopts an
// existing device account with the same username.
func (s *RegistrationService) Install(ctx context.Context, in MarkInstalledInput) (MarkInstalledResult, error) {
	reg, err := s.regs.FindByID(ctx, in.ID)
	if err != nil {
		return MarkInstalledResult{}, err
	}
	if reg.Status != domainReg.StatusApproved {
		return MarkInstalledResult{}, fmt.Errorf("%w: expected APPROVED, got %s", domainReg.ErrInvalidTransition, reg.Status)
	}
	if in.DeviceID == "" && reg.DeviceID != "" {
		in.DeviceID = reg.DeviceID
	}
	if in.DeviceID == "" {
		return MarkInstalledResult{}, domainReg.ErrDeviceRequired
	}
	planRow, err := s.plans.FindByID(ctx, reg.PlanID)
	if err != nil {
		return MarkInstalledResult{}, fmt.Errorf("load plan: %w", err)
	}

	customerID := reg.CustomerID
	if customerID == "" {
		custID, err := s.upsertCustomer(ctx, reg)
		if err != nil {
			return MarkInstalledResult{}, err
		}
		customerID = custID
	}

	subID := reg.SubscriptionID
	username := normalizeUsername(in.RemoteUsername)
	if username == "" {
		username = normalizeUsername(reg.FullName)
	}
	password := in.Password

	var sub domainSub.Subscription
	if subID != "" {
		sub, err = s.loadSubscriptionByID(ctx, subID)
	} else {
		// Idempotent username: adopt or disambiguate against this device.
		existing, lookupErr := s.subs.FindByDeviceAndUsername(ctx, in.DeviceID, username)
		switch {
		case lookupErr == nil && existing.ID != "" && existing.CustomerID == "":
			_ = existing // orphan mapping — fall through to generate a unique name
		case lookupErr == nil:
			suffix, gerr := randomSuffix()
			if gerr != nil {
				return MarkInstalledResult{}, gerr
			}
			username = username + "-" + suffix
		}
		sub = domainSub.Subscription{
			ID:             newID(),
			TenantID:       reg.TenantID,
			CustomerID:     customerID,
			PlanID:         planRow.ID,
			DeviceID:       in.DeviceID,
			ServiceType:    planRow.ServiceType,
			RemoteUsername: username,
			BillingDay:     1,
			Status:         domainSub.StatusPendingProvision,
			StartDate:      time.Now(),
		}
		err = s.subs.Save(ctx, sub)
	}
	if err != nil {
		return MarkInstalledResult{}, fmt.Errorf("persist subscription: %w", err)
	}
	if sub.RemoteUsername != "" && username == "" {
		username = sub.RemoteUsername
	}
	sub.RemoteUsername = username
	sub.DeviceID = in.DeviceID
	if password == "" {
		password, err = generatePassword()
		if err != nil {
			return MarkInstalledResult{}, fmt.Errorf("generate password: %w", err)
		}
	}

	// Device-side provisioning; sets ACTIVE + mapping on success.
	sub, err = s.provisioner.Provision(ctx, sub, password)
	if err != nil {
		// Subscription stays PENDING_PROVISION for a safe retry.
		return MarkInstalledResult{}, fmt.Errorf("provision subscription: %w", err)
	}

	now := time.Now()
	reg.Status = domainReg.StatusActive
	reg.DeviceID = in.DeviceID
	reg.CustomerID = customerID
	reg.SubscriptionID = sub.ID
	reg.InstalledAt = &now
	reg.TechnicianNotes = in.TechnicianNotes
	if err := s.regs.Save(ctx, reg); err != nil {
		return MarkInstalledResult{}, fmt.Errorf("finalize registration: %w", err)
	}

	logger.WithComponent("Registration").WithFields(map[string]any{
		"registration_id": reg.ID, "customer_id": customerID,
		"subscription_id": sub.ID, "device_id": in.DeviceID,
	}).Info("registration installed & provisioned")
	return MarkInstalledResult{Registration: reg, CustomerID: customerID, SubscriptionID: sub.ID}, nil
}

// Cancel terminates a PENDING/APPROVED registration.
func (s *RegistrationService) Cancel(ctx context.Context, id, reason string) (domainReg.Registration, error) {
	reg, err := s.regs.FindByID(ctx, id)
	if err != nil {
		return domainReg.Registration{}, err
	}
	if reg.Status != domainReg.StatusPending && reg.Status != domainReg.StatusApproved {
		return reg, fmt.Errorf("%w: %s cannot be cancelled", domainReg.ErrInvalidTransition, reg.Status)
	}
	now := time.Now()
	reg.Status = domainReg.StatusCancelled
	reg.CancelledAt = &now
	reg.CancelReason = reason
	if err := s.regs.Save(ctx, reg); err != nil {
		return reg, fmt.Errorf("save cancellation: %w", err)
	}
	logger.WithComponent("Registration").WithField("registration_id", id).
		Info("registration cancelled")
	return reg, nil
}

// upsertCustomer returns the existing active customer for this phone, or
// creates one carrying over the registration's intake data.
func (s *RegistrationService) upsertCustomer(ctx context.Context, reg domainReg.Registration) (string, error) {
	if existing, err := s.customers.FindByPhone(ctx, reg.Phone); err == nil && existing.ID != "" {
		return existing.ID, nil
	}
	code, err := s.customers.NextCustomerCode(ctx)
	if err != nil {
		return "", fmt.Errorf("generate customer code: %w", err)
	}
	cust := domainCustomer.Customer{
		ID:           newID(),
		TenantID:     reg.TenantID,
		CustomerCode: code,
		Name:         reg.FullName,
		Phone:        reg.Phone,
		Address:      reg.Address,
		Latitude:     reg.Latitude,
		Longitude:    reg.Longitude,
		Status:       "ACTIVE",
		Notes:        reg.Notes,
	}
	if err := s.customers.Save(ctx, cust); err != nil {
		return "", fmt.Errorf("create customer: %w", err)
	}
	return cust.ID, nil
}

func (s *RegistrationService) loadSubscriptionByID(ctx context.Context, id string) (domainSub.Subscription, error) {
	finder, ok := s.subs.(interface {
		FindByID(context.Context, string) (domainSub.Subscription, error)
	})
	if !ok {
		return domainSub.Subscription{}, fmt.Errorf("subscription repository does not support FindByID")
	}
	return finder.FindByID(ctx, id)
}

// normalizeUsername keeps [A-Z0-9-_]; RouterOS usernames are case-sensitive
// but ISP convention is uppercase like the legacy exports ("MATRAJI-KT").
func normalizeUsername(name string) string {
	var b strings.Builder
	for _, r := range name {
		switch {
		case r >= 'a' && r <= 'z':
			b.WriteRune(r - 32)
		case r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '-', r == '_', r == '.':
			b.WriteRune(r)
		case r == ' ':
			b.WriteByte('-')
		}
	}
	return strings.Trim(b.String(), "-")
}

const passwordAlphabet = "abcdefghjkmnpqrstuvwxyzABCDEFGHJKMNPQRSTUVWXYZ23456789"

func generatePassword() (string, error) {
	out := make([]byte, 8)
	for i := range out {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(passwordAlphabet))))
		if err != nil {
			return "", err
		}
		out[i] = passwordAlphabet[n.Int64()]
	}
	return string(out), nil
}

func randomSuffix() (string, error) {
	const chars = "0123456789abcdef"
	out := make([]byte, 4)
	for i := range out {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		if err != nil {
			return "", err
		}
		out[i] = chars[n.Int64()]
	}
	return string(out), nil
}
