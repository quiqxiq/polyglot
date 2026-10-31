package olt

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/quixiq/polyglot/internal/domain/routerjob"
	"github.com/quixiq/polyglot/pkg/idgen"
)

// EnqueueRename menjadwalkan rename ONU (aksi tulis dimulai dari sini).
func (s *Service) EnqueueRename(ctx context.Context, oltDeviceID string, ponIndex, onuIndex int, name string) (routerjob.Job, error) {
	return s.enqueue(ctx, oltDeviceID, routerjob.TypeONTRename, map[string]any{
		"olt_device_id": oltDeviceID, "pon_index": ponIndex, "onu_index": onuIndex, "name": name,
	}, fmt.Sprintf("olt:%s:ont-rename:%d/%d:%s", oltDeviceID, ponIndex, onuIndex, name))
}

// EnqueueReboot menjadwalkan reboot ONU.
func (s *Service) EnqueueReboot(ctx context.Context, oltDeviceID string, ponIndex, onuIndex int) (routerjob.Job, error) {
	return s.enqueue(ctx, oltDeviceID, routerjob.TypeONTReboot, map[string]any{
		"olt_device_id": oltDeviceID, "pon_index": ponIndex, "onu_index": onuIndex,
	}, fmt.Sprintf("olt:%s:ont-reboot:%d/%d:%s", oltDeviceID, ponIndex, onuIndex, minuteBucket()))
}

// EnqueueSetAdmin menjadwalkan enable/disable ONU.
func (s *Service) EnqueueSetAdmin(ctx context.Context, oltDeviceID string, ponIndex, onuIndex int, enabled bool) (routerjob.Job, error) {
	return s.enqueue(ctx, oltDeviceID, routerjob.TypeONTSetAdmin, map[string]any{
		"olt_device_id": oltDeviceID, "pon_index": ponIndex, "onu_index": onuIndex, "enabled": enabled,
	}, fmt.Sprintf("olt:%s:ont-admin:%d/%d:%t:%s", oltDeviceID, ponIndex, onuIndex, enabled, minuteBucket()))
}

// EnqueueDelete menjadwalkan hapus ONU dari OLT.
func (s *Service) EnqueueDelete(ctx context.Context, oltDeviceID string, ponIndex, onuIndex int) (routerjob.Job, error) {
	return s.enqueue(ctx, oltDeviceID, routerjob.TypeONTDelete, map[string]any{
		"olt_device_id": oltDeviceID, "pon_index": ponIndex, "onu_index": onuIndex,
	}, fmt.Sprintf("olt:%s:ont-delete:%d/%d", oltDeviceID, ponIndex, onuIndex))
}

// EnqueueSync menjadwalkan sync inventori ONU.
func (s *Service) EnqueueSync(ctx context.Context, oltDeviceID string) (routerjob.Job, error) {
	return s.enqueue(ctx, oltDeviceID, routerjob.TypeSyncOLT, map[string]any{
		"olt_device_id": oltDeviceID,
	}, fmt.Sprintf("olt:%s:sync:%s", oltDeviceID, minuteBucket()))
}

// EnqueueSaveConfig menjadwalkan simpan konfigurasi OLT ke NVRAM.
func (s *Service) EnqueueSaveConfig(ctx context.Context, oltDeviceID string) (routerjob.Job, error) {
	return s.enqueue(ctx, oltDeviceID, routerjob.TypeOLTSaveConfig, map[string]any{
		"olt_device_id": oltDeviceID,
	}, fmt.Sprintf("olt:%s:save-config:%s", oltDeviceID, minuteBucket()))
}

func (s *Service) enqueue(ctx context.Context, oltDeviceID string, typ routerjob.Type, payload map[string]any, idempotencyKey string) (routerjob.Job, error) {
	if s.deps.Jobs == nil {
		return routerjob.Job{}, fmt.Errorf("olt: job queue belum dikonfigurasi")
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return routerjob.Job{}, fmt.Errorf("olt: marshal payload: %w", err)
	}
	now := time.Now().UTC()
	job := routerjob.Job{
		ID:             idgen.New("rj"),
		TenantID:       tenantDefault,
		DeviceID:       oltDeviceID,
		JobType:        typ,
		Payload:        raw,
		IdempotencyKey: idempotencyKey,
		Status:         routerjob.StatusQueued,
		MaxAttempts:    5,
		NextRunAt:      now,
	}
	return s.deps.Jobs.Enqueue(ctx, job)
}

// --- aksi langsung (dipanggil worker job) ---

// RenameONU mengeksekusi rename ONU ke OLT.
func (s *Service) RenameONU(ctx context.Context, oltDeviceID string, ponIndex, onuIndex int, name string) error {
	mgr, err := s.deps.Managers.ForDevice(ctx, oltDeviceID)
	if err != nil {
		return err
	}
	if err := mgr.RenameONU(ctx, ponIndex, onuIndex, name); err != nil {
		return err
	}
	s.recordONTEvent(ctx, oltDeviceID, ponIndex, onuIndex, "RENAMED", "ONU di-rename menjadi "+name)
	return nil
}

// RebootONU mengeksekusi reboot ONU.
func (s *Service) RebootONU(ctx context.Context, oltDeviceID string, ponIndex, onuIndex int) error {
	mgr, err := s.deps.Managers.ForDevice(ctx, oltDeviceID)
	if err != nil {
		return err
	}
	if err := mgr.RebootONU(ctx, ponIndex, onuIndex); err != nil {
		return err
	}
	s.recordONTEvent(ctx, oltDeviceID, ponIndex, onuIndex, "REBOOTED", "reboot ONU dikirim")
	return nil
}

// SetONUAdmin mengeksekusi enable/disable ONU.
func (s *Service) SetONUAdmin(ctx context.Context, oltDeviceID string, ponIndex, onuIndex int, enabled bool) error {
	mgr, err := s.deps.Managers.ForDevice(ctx, oltDeviceID)
	if err != nil {
		return err
	}
	if err := mgr.SetONUAdmin(ctx, ponIndex, onuIndex, enabled); err != nil {
		return err
	}
	state := "DISABLED"
	if enabled {
		state = "ENABLED"
	}
	s.recordONTEvent(ctx, oltDeviceID, ponIndex, onuIndex, state, "status administrasi ONU diubah")
	return nil
}

// DeleteONU mengeksekusi hapus ONU.
func (s *Service) DeleteONU(ctx context.Context, oltDeviceID string, ponIndex, onuIndex int) error {
	mgr, err := s.deps.Managers.ForDevice(ctx, oltDeviceID)
	if err != nil {
		return err
	}
	if err := mgr.DeleteONU(ctx, ponIndex, onuIndex); err != nil {
		return err
	}
	if o, errFind := s.deps.ONTs.FindByOLTPosition(ctx, oltDeviceID, ponIndex, onuIndex); errFind == nil {
		o.State = "REMOVED"
		_ = s.deps.ONTs.Save(ctx, o)
	}
	s.recordONTEvent(ctx, oltDeviceID, ponIndex, onuIndex, "DELETED", "ONU dihapus dari OLT")
	return nil
}

// SaveOLTConfig menyimpan konfigurasi OLT ke NVRAM.
func (s *Service) SaveOLTConfig(ctx context.Context, oltDeviceID string) error {
	mgr, err := s.deps.Managers.ForDevice(ctx, oltDeviceID)
	if err != nil {
		return err
	}
	return mgr.SaveConfig(ctx)
}

func (s *Service) recordONTEvent(ctx context.Context, oltDeviceID string, ponIndex, onuIndex int, eventType, message string) {
	o, err := s.deps.ONTs.FindByOLTPosition(ctx, oltDeviceID, ponIndex, onuIndex)
	if err != nil {
		return
	}
	s.recordEvent(ctx, o, eventType, "", "", message)
}

func minuteBucket() string {
	return time.Now().UTC().Format("200601021504")
}
