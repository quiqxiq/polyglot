package routerjob

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	domainRouterJob "github.com/quixiq/polyglot/internal/domain/routerjob"
	domainSubscription "github.com/quixiq/polyglot/internal/domain/subscription"
	"github.com/quixiq/polyglot/internal/port"
)

// Exporter menghasilkan file script RouterOS (.rsc) manual sebagai fallback
// bila koneksi ke router tidak stabil (planning v2 §1.13 / analisis §3.1.B).
type Exporter struct {
	jobs  port.RouterJobRepository
	subs  port.SubscriptionRepository
	plans port.ServicePlanRepository
}

// NewExporter wires the .rsc exporter.
func NewExporter(jobs port.RouterJobRepository, subs port.SubscriptionRepository, plans port.ServicePlanRepository) *Exporter {
	return &Exporter{jobs: jobs, subs: subs, plans: plans}
}

// ExportScript mengembalikan isi file .rsc untuk satu job.
func (e *Exporter) ExportScript(ctx context.Context, jobID string) (string, error) {
	job, err := e.jobs.FindByID(ctx, jobID)
	if err != nil {
		return "", err
	}
	var lines []string
	lines = append(lines, "# Polyglot router fallback script")
	lines = append(lines, fmt.Sprintf("# Job      : %s", job.ID))
	lines = append(lines, fmt.Sprintf("# Type     : %s", job.JobType))
	lines = append(lines, fmt.Sprintf("# Device   : %s", job.DeviceID))
	lines = append(lines, fmt.Sprintf("# Status   : %s (attempts %d/%d)", job.Status, job.Attempts, job.MaxAttempts))
	lines = append(lines, fmt.Sprintf("# Generated: %s", time.Now().UTC().Format(time.RFC3339)))
	if job.LastError != "" {
		lines = append(lines, "# Last error: "+strings.ReplaceAll(job.LastError, "\n", " "))
	}
	lines = append(lines, "")

	switch job.JobType {
	case domainRouterJob.TypeProvisionPPPoE, domainRouterJob.TypeProvisionHotspot,
		domainRouterJob.TypeProvisionDedicated:
		subID, perr := payloadSubscriptionID(job)
		if perr != nil {
			return "", perr
		}
		sub, serr := e.subs.FindByID(ctx, subID)
		if serr != nil {
			return "", fmt.Errorf("subscription %s: %w", subID, serr)
		}
		pl, plerr := e.plans.FindByID(ctx, sub.PlanID)
		if plerr != nil {
			return "", fmt.Errorf("plan %s: %w", sub.PlanID, plerr)
		}
		lines = append(lines, e.provisionScript(sub, pl.RouterProfile())...)
	case domainRouterJob.TypeIsolate, domainRouterJob.TypeRestore, domainRouterJob.TypeDeleteAccount:
		subID, perr := payloadSubscriptionID(job)
		if perr != nil {
			return "", perr
		}
		sub, serr := e.subs.FindByID(ctx, subID)
		if serr != nil {
			return "", fmt.Errorf("subscription %s: %w", subID, serr)
		}
		switch job.JobType {
		case domainRouterJob.TypeIsolate:
			lines = append(lines, e.isolateScript(sub)...)
		case domainRouterJob.TypeRestore:
			lines = append(lines, e.restoreScript(sub)...)
		default:
			lines = append(lines, e.deleteScript(sub)...)
		}
	case domainRouterJob.TypeSyncProfile:
		var payload struct {
			PlanID string `json:"plan_id"`
		}
		if err := json.Unmarshal(job.Payload, &payload); err != nil || payload.PlanID == "" {
			return "", fmt.Errorf("%w: payload plan_id invalid", domainRouterJob.ErrInvalidInput)
		}
		pl, plerr := e.plans.FindByID(ctx, payload.PlanID)
		if plerr != nil {
			return "", fmt.Errorf("plan %s: %w", payload.PlanID, plerr)
		}
		lines = append(lines, fmt.Sprintf(
			"/ppp profile add name=%q rate-limit=%q comment=%q",
			pl.RouterProfile(), pl.RateLimitWithBurst(), "polyglot plan:"+pl.ID,
		))
	default:
		return "", fmt.Errorf("%w: job type %s does not support script export", domainRouterJob.ErrUnsupportedType, job.JobType)
	}
	return strings.Join(lines, "\n") + "\n", nil
}

func (e *Exporter) provisionScript(sub domainSubscription.Subscription, profile string) []string {
	switch strings.ToLower(sub.ServiceType) {
	case "hotspot":
		return []string{fmt.Sprintf(
			"/ip hotspot user add name=%q password=%q profile=%q comment=%q",
			sub.RemoteUsername, sub.RemotePassword, profile, "polyglot:"+sub.ID,
		)}
	case "dedicated":
		target := ""
		if sub.RemoteAddress != "" {
			target = sub.RemoteAddress + "/32"
		}
		return []string{fmt.Sprintf(
			"/queue simple add name=%q target=%q max-limit=%q comment=%q",
			sub.RemoteUsername, target, sub.RateLimit, "polyglot:"+sub.ID,
		)}
	default:
		return []string{fmt.Sprintf(
			"/ppp secret add name=%q password=%q profile=%q service=pppoe comment=%q",
			sub.RemoteUsername, sub.RemotePassword, profile, "polyglot:"+sub.ID,
		)}
	}
}

func (e *Exporter) isolateScript(sub domainSubscription.Subscription) []string {
	return []string{
		fmt.Sprintf("/ip firewall address-list add address=%q list=%q comment=%q",
			sub.RemoteAddress, "ISOLIR_USERS", "isolir:"+sub.RemoteUsername),
		fmt.Sprintf("/ppp secret set [find name=%q] profile=%q", sub.RemoteUsername, "isolir"),
		fmt.Sprintf("/ppp active remove [find name=%q]", sub.RemoteUsername),
	}
}

func (e *Exporter) restoreScript(sub domainSubscription.Subscription) []string {
	profile := sub.RouterProfile
	if profile == "" {
		profile = "default"
	}
	return []string{
		fmt.Sprintf("/ip firewall address-list remove [find comment=%q]", "isolir:"+sub.RemoteUsername),
		fmt.Sprintf("/ppp secret set [find name=%q] profile=%q", sub.RemoteUsername, profile),
		fmt.Sprintf("/ppp active remove [find name=%q]", sub.RemoteUsername),
	}
}

func (e *Exporter) deleteScript(sub domainSubscription.Subscription) []string {
	if strings.EqualFold(sub.ServiceType, "HOTSPOT") {
		return []string{fmt.Sprintf("/ip hotspot user remove [find name=%q]", sub.RemoteUsername)}
	}
	return []string{
		fmt.Sprintf("/ppp secret remove [find name=%q]", sub.RemoteUsername),
		fmt.Sprintf("/ip firewall address-list remove [find comment=%q]", "isolir:"+sub.RemoteUsername),
	}
}
