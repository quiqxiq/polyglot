package registration

import (
	"context"
	"fmt"
	"time"

	domainReg "github.com/quixiq/polyglot/internal/domain/registration"
	"github.com/quixiq/polyglot/pkg/logger"
)

// Review approves or rejects a PENDING registration and optionally records
// the installation schedule + assigned technician.
func (s *RegistrationService) Review(ctx context.Context, id string, approve bool, reviewedBy int64, notes string, scheduledInstall *time.Time, technicianID int64) (domainReg.Registration, error) {
	reg, err := s.regs.FindByID(ctx, id)
	if err != nil {
		return domainReg.Registration{}, err
	}
	if reg.Status != domainReg.StatusPending {
		return reg, fmt.Errorf("%w: expected PENDING, got %s", domainReg.ErrInvalidTransition, reg.Status)
	}

	now := time.Now()
	reg.ReviewedBy = &reviewedBy
	reg.ReviewedAt = &now

	if !approve {
		reg.Status = domainReg.StatusRejected
		reg.RejectedAt = &now
		reg.RejectedReason = notes
		if err := s.regs.Save(ctx, reg); err != nil {
			return reg, fmt.Errorf("save rejection: %w", err)
		}
		logger.WithComponent("Registration").WithField("registration_id", id).
			Info("registration rejected")
		return reg, nil
	}

	reg.Status = domainReg.StatusApproved
	reg.AdminNotes = notes
	if scheduledInstall != nil {
		reg.ScheduledInstallDate = scheduledInstall
	}
	if technicianID > 0 {
		tid := technicianID
		reg.AssignedTechnicianID = &tid
	}

	if err := s.regs.Save(ctx, reg); err != nil {
		return reg, fmt.Errorf("save approval: %w", err)
	}
	logger.WithComponent("Registration").WithFields(map[string]any{
		"registration_id": id, "reviewed_by": reviewedBy,
	}).Info("registration approved")
	return reg, nil
}
