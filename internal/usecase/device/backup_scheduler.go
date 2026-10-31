package device

import (
	"context"
	"fmt"
	"time"

	"github.com/quixiq/polyglot/internal/port"
	networkUC "github.com/quixiq/polyglot/internal/usecase/network"
)

// BackupScheduler runs daily/weekly restore-point backups for enabled
// MikroTik devices. Advisory locking is handled by the app scheduler; this
// usecase only iterates devices and skips devices with an active approval.
type BackupScheduler struct {
	Devices   port.DeviceRepository
	Restore   *networkUC.RestorePointUseCase
	Approvals port.ApprovalRepository
	TenantID  string
	Now       func() time.Time
}

// Run creates one backup per eligible device. It returns created/failed and
// never aborts the whole batch for one device failure.
func (s *BackupScheduler) Run(ctx context.Context) (created, failed int, err error) {
	if s.Restore == nil || s.Devices == nil {
		return 0, 0, fmt.Errorf("backup scheduler dependencies not configured")
	}
	if s.Now == nil {
		s.Now = func() time.Time { return time.Now().UTC() }
	}
	tenant := s.TenantID
	if tenant == "" {
		tenant = "tenant-default"
	}
	devs, derr := s.Devices.FindAll(ctx)
	if derr != nil {
		return 0, 0, derr
	}
	for _, dev := range devs {
		if !dev.Enabled || dev.DriverType != "mikrotik" {
			continue
		}
		if s.hasActiveApproval(ctx, tenant, dev.ID) {
			continue
		}
		if _, rerr := s.Restore.CreateScheduled(ctx, tenant, dev.ID); rerr != nil {
			failed++
			continue
		}
		created++
	}
	return created, failed, nil
}

func (s *BackupScheduler) hasActiveApproval(ctx context.Context, tenantID, deviceID string) bool {
	if s.Approvals == nil {
		return false
	}
	for _, status := range []string{"PENDING", "APPROVED", "EXECUTING"} {
		items, _, err := s.Approvals.List(ctx, port.ApprovalFilter{TenantID: tenantID, DeviceID: deviceID, Status: status, Limit: 1})
		if err == nil && len(items) > 0 {
			return true
		}
	}
	return false
}

// NewBackupScheduler builds the scheduled-backup usecase.
func NewBackupScheduler(cfg BackupScheduler) *BackupScheduler {
	if cfg.Now == nil {
		cfg.Now = func() time.Time { return time.Now().UTC() }
	}
	if cfg.TenantID == "" {
		cfg.TenantID = "tenant-default"
	}
	return &cfg
}
