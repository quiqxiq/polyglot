package routerjob

import (
	"context"
	"encoding/json"
	"fmt"

	domainRouterJob "github.com/quixiq/polyglot/internal/domain/routerjob"
)

// ontPositionPayload adalah payload job aksi ONU pada OLT.
type ontPositionPayload struct {
	OLTDeviceID string `json:"olt_device_id"`
	PONIndex    int    `json:"pon_index"`
	ONUIndex    int    `json:"onu_index"`
	Name        string `json:"name,omitempty"`
	Enabled     bool   `json:"enabled,omitempty"`
}

func parseONTPosition(job domainRouterJob.Job) (ontPositionPayload, error) {
	var p ontPositionPayload
	if err := json.Unmarshal(job.Payload, &p); err != nil {
		return p, fmt.Errorf("payload job %s tidak valid: %w", job.JobType, err)
	}
	if p.OLTDeviceID == "" || p.PONIndex <= 0 || p.ONUIndex <= 0 {
		return p, fmt.Errorf("payload job %s tidak lengkap", job.JobType)
	}
	return p, nil
}

func oltRequired(deps HandlerDeps) error {
	if deps.OLT == nil {
		return fmt.Errorf("OLT actions belum dikonfigurasi")
	}
	return nil
}

func syncOLTHandler(deps HandlerDeps) Handler {
	return func(ctx context.Context, job domainRouterJob.Job) error {
		if err := oltRequired(deps); err != nil {
			return err
		}
		var p struct {
			OLTDeviceID string `json:"olt_device_id"`
		}
		if err := json.Unmarshal(job.Payload, &p); err != nil || p.OLTDeviceID == "" {
			return fmt.Errorf("payload SYNC_OLT tidak valid")
		}
		_, err := deps.OLT.SyncONUs(ctx, p.OLTDeviceID)
		return err
	}
}

func ontRenameHandler(deps HandlerDeps) Handler {
	return func(ctx context.Context, job domainRouterJob.Job) error {
		if err := oltRequired(deps); err != nil {
			return err
		}
		p, err := parseONTPosition(job)
		if err != nil {
			return err
		}
		return deps.OLT.RenameONU(ctx, p.OLTDeviceID, p.PONIndex, p.ONUIndex, p.Name)
	}
}

func ontRebootHandler(deps HandlerDeps) Handler {
	return func(ctx context.Context, job domainRouterJob.Job) error {
		if err := oltRequired(deps); err != nil {
			return err
		}
		p, err := parseONTPosition(job)
		if err != nil {
			return err
		}
		return deps.OLT.RebootONU(ctx, p.OLTDeviceID, p.PONIndex, p.ONUIndex)
	}
}

func ontSetAdminHandler(deps HandlerDeps) Handler {
	return func(ctx context.Context, job domainRouterJob.Job) error {
		if err := oltRequired(deps); err != nil {
			return err
		}
		p, err := parseONTPosition(job)
		if err != nil {
			return err
		}
		return deps.OLT.SetONUAdmin(ctx, p.OLTDeviceID, p.PONIndex, p.ONUIndex, p.Enabled)
	}
}

func ontDeleteHandler(deps HandlerDeps) Handler {
	return func(ctx context.Context, job domainRouterJob.Job) error {
		if err := oltRequired(deps); err != nil {
			return err
		}
		p, err := parseONTPosition(job)
		if err != nil {
			return err
		}
		return deps.OLT.DeleteONU(ctx, p.OLTDeviceID, p.PONIndex, p.ONUIndex)
	}
}

func oltSaveConfigHandler(deps HandlerDeps) Handler {
	return func(ctx context.Context, job domainRouterJob.Job) error {
		if err := oltRequired(deps); err != nil {
			return err
		}
		var p struct {
			OLTDeviceID string `json:"olt_device_id"`
		}
		if err := json.Unmarshal(job.Payload, &p); err != nil || p.OLTDeviceID == "" {
			return fmt.Errorf("payload OLT_SAVE_CONFIG tidak valid")
		}
		return deps.OLT.SaveOLTConfig(ctx, p.OLTDeviceID)
	}
}
