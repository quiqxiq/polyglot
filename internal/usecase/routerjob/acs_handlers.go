package routerjob

import (
	"context"
	"encoding/json"
	"fmt"

	domainRouterJob "github.com/quixiq/polyglot/internal/domain/routerjob"
)

type acsOntPayload struct {
	ONTID    string `json:"ont_id"`
	SSID     string `json:"ssid,omitempty"`
	Password string `json:"password,omitempty"`
	Username string `json:"username,omitempty"`
	Band     string `json:"band,omitempty"`
}

func parseACSJob(job domainRouterJob.Job) (acsOntPayload, error) {
	var p acsOntPayload
	if err := json.Unmarshal(job.Payload, &p); err != nil {
		return p, fmt.Errorf("payload job %s tidak valid: %w", job.JobType, err)
	}
	if p.ONTID == "" && job.JobType != domainRouterJob.TypeACSSyncMatch && job.JobType != domainRouterJob.TypeACSBootstrapPreset {
		return p, fmt.Errorf("payload job %s tidak memiliki ont_id", job.JobType)
	}
	return p, nil
}

func acsRequired(deps HandlerDeps) error {
	if deps.ACS == nil {
		return fmt.Errorf("ACS actions belum dikonfigurasi")
	}
	return nil
}

func acsSyncMatchHandler(deps HandlerDeps) Handler {
	return func(ctx context.Context, _ domainRouterJob.Job) error {
		if err := acsRequired(deps); err != nil {
			return err
		}
		_, err := deps.ACS.SyncDevices(ctx)
		return err
	}
}

func acsRefreshTelemetryHandler(deps HandlerDeps) Handler {
	return func(ctx context.Context, job domainRouterJob.Job) error {
		if err := acsRequired(deps); err != nil {
			return err
		}
		p, err := parseACSJob(job)
		if err != nil {
			return err
		}
		return deps.ACS.RefreshTelemetry(ctx, p.ONTID)
	}
}

func acsRebootHandler(deps HandlerDeps) Handler {
	return func(ctx context.Context, job domainRouterJob.Job) error {
		if err := acsRequired(deps); err != nil {
			return err
		}
		p, err := parseACSJob(job)
		if err != nil {
			return err
		}
		return deps.ACS.RebootDevice(ctx, p.ONTID)
	}
}

func acsSetWiFiHandler(deps HandlerDeps) Handler {
	return func(ctx context.Context, job domainRouterJob.Job) error {
		if err := acsRequired(deps); err != nil {
			return err
		}
		p, err := parseACSJob(job)
		if err != nil {
			return err
		}
		return deps.ACS.SetWiFi(ctx, p.ONTID, p.SSID, p.Password, p.Band)
	}
}

func acsPushPPPoEHandler(deps HandlerDeps) Handler {
	return func(ctx context.Context, job domainRouterJob.Job) error {
		if err := acsRequired(deps); err != nil {
			return err
		}
		p, err := parseACSJob(job)
		if err != nil {
			return err
		}
		return deps.ACS.PushPPPoE(ctx, p.ONTID, p.Username, p.Password)
	}
}

func acsBootstrapPresetHandler(deps HandlerDeps) Handler {
	return func(ctx context.Context, _ domainRouterJob.Job) error {
		if err := acsRequired(deps); err != nil {
			return err
		}
		return deps.ACS.BootstrapPreset(ctx)
	}
}

func ontBindHandler(_ HandlerDeps) Handler {
	return func(_ context.Context, job domainRouterJob.Job) error {
		// Fase E: binding PSB → ONT (rename/link) diimplementasikan pada ronde
		// integrasi registrasi; untuk sekarang job ini no-op terkontrol.
		return nil
	}
}
