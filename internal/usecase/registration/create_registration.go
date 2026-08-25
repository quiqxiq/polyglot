package registration

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	domainCustomer "github.com/quixiq/polyglot/internal/domain/customer"
	domainPlan "github.com/quixiq/polyglot/internal/domain/plan"
	domainReg "github.com/quixiq/polyglot/internal/domain/registration"
	domainSub "github.com/quixiq/polyglot/internal/domain/subscription"
	"github.com/quixiq/polyglot/pkg/logger"
)

// RegistrationRepository narrows persistence for the workflow.
type RegistrationRepository interface {
	Save(ctx context.Context, r domainReg.Registration) error
	FindByID(ctx context.Context, id string) (domainReg.Registration, error)
	List(ctx context.Context, status string, limit int) ([]domainReg.Registration, error)
	HasActiveByPhone(ctx context.Context, phone string) (bool, error)
}

// Narrow persistence seams so unit tests stub exactly what each step touches.
type planLookup interface {
	FindByID(ctx context.Context, id string) (domainPlan.Plan, error)
}

type customerWriter interface {
	Save(ctx context.Context, c domainCustomer.Customer) error
	FindByPhone(ctx context.Context, phone string) (domainCustomer.Customer, error)
	NextCustomerCode(ctx context.Context) (string, error)
}

type subscriptionWriter interface {
	Save(ctx context.Context, s domainSub.Subscription) error
	FindByDeviceAndUsername(ctx context.Context, deviceID, username string) (domainSub.Subscription, error)
}

// Provisioner is the slice of network.SubscriptionProvisioner needed by the
// install step; declared locally to keep the dependency direction explicit.
type Provisioner interface {
	Provision(ctx context.Context, sub domainSub.Subscription, password string) (domainSub.Subscription, error)
}

// RegistrationService implements the intake → review → install workflow.
// MarkInstalled (install_registration.go) triggers provisioning through the
// Provisioner; this file holds shared wiring and the intake operations.
type RegistrationService struct {
	regs        RegistrationRepository
	plans       planLookup
	customers   customerWriter
	subs        subscriptionWriter
	provisioner Provisioner
}

func NewRegistrationService(
	regs RegistrationRepository,
	plans planLookup,
	customers customerWriter,
	subs subscriptionWriter,
	provisioner Provisioner,
) *RegistrationService {
	return &RegistrationService{
		regs: regs, plans: plans, customers: customers,
		subs: subs, provisioner: provisioner,
	}
}

// Create validates the concise application form and stores it as PENDING.
func (s *RegistrationService) Create(ctx context.Context, reg domainReg.Registration) (domainReg.Registration, error) {
	if err := reg.Validate(); err != nil {
		return domainReg.Registration{}, err
	}
	if active, err := s.regs.HasActiveByPhone(ctx, reg.Phone); err != nil {
		return domainReg.Registration{}, fmt.Errorf("check duplicate: %w", err)
	} else if active {
		return domainReg.Registration{}, domainReg.ErrAlreadyPending
	}
	if _, err := s.plans.FindByID(ctx, reg.PlanID); err != nil {
		return domainReg.Registration{}, fmt.Errorf("plan not found: %w", err)
	}

	if reg.ID == "" {
		reg.ID = uuid.NewString()
	}
	if reg.TenantID == "" {
		reg.TenantID = "tenant-default"
	}
	if reg.Status == "" {
		reg.Status = domainReg.StatusPending
	}
	if reg.RegistrationNo == "" {
		reg.RegistrationNo = generateRegistrationNo()
	}
	now := time.Now()
	if reg.CreatedAt.IsZero() {
		reg.CreatedAt = now
	}
	if err := s.regs.Save(ctx, reg); err != nil {
		return domainReg.Registration{}, fmt.Errorf("save registration: %w", err)
	}
	logger.WithComponent("Registration").WithFields(map[string]any{
		"registration_id": reg.ID, "no": reg.RegistrationNo, "plan_id": reg.PlanID,
	}).Info("registration created")
	return reg, nil
}

func (s *RegistrationService) Get(ctx context.Context, id string) (domainReg.Registration, error) {
	if id == "" {
		return domainReg.Registration{}, domainReg.ErrNotFound
	}
	return s.regs.FindByID(ctx, id)
}

func (s *RegistrationService) List(ctx context.Context, status string, limit int) ([]domainReg.Registration, error) {
	switch status {
	case "", domainReg.StatusPending, domainReg.StatusApproved, domainReg.StatusInstalled,
		domainReg.StatusActive, domainReg.StatusRejected, domainReg.StatusCancelled:
	default:
		return nil, domainReg.ErrInvalidTransition
	}
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	return s.regs.List(ctx, status, limit)
}

// generateRegistrationNo produces REG-YYYYMM-XXXX.
func generateRegistrationNo() string {
	now := time.Now().UTC()
	tail := uuid.NewString()[0:4]
	return fmt.Sprintf("REG-%04d%02d-%s", now.Year(), int(now.Month()), tail)
}
