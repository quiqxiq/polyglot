// Package routerjob menjalankan antrean job asynchronous per perangkat router
// (planning v2 Fase 4): batch 20 job/step, error budget 5 error beruntun,
// retry exponential backoff, status DEAD, dan retry manual dari dashboard.
package routerjob

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	domainRouterJob "github.com/quixiq/polyglot/internal/domain/routerjob"
	domainTenant "github.com/quixiq/polyglot/internal/domain/tenant"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/pkg/idgen"
	"github.com/quixiq/polyglot/pkg/logger"
)

// Handler mengeksekusi satu job router.
type Handler func(ctx context.Context, job domainRouterJob.Job) error

// Result merangkum satu siklus worker.
type Result struct {
	Processed    int
	Succeeded    int
	Retried      int
	Dead         int
	StoppedEarly bool
}

// Worker memproses antrean router_jobs secara periodik.
type Worker struct {
	jobs        port.RouterJobRepository
	handlers    map[domainRouterJob.Type]Handler
	settings    port.SettingReader
	batchSize   int
	errorBudget int

	now func() time.Time
}

// NewWorker wires the worker with a handler per job type.
func NewWorker(
	jobs port.RouterJobRepository,
	handlers map[domainRouterJob.Type]Handler,
	settings port.SettingReader,
) *Worker {
	return &Worker{
		jobs: jobs, handlers: handlers, settings: settings,
		batchSize: 20, errorBudget: 5, now: time.Now,
	}
}

// WithClock mengganti sumber waktu (untuk test deterministik).
func (w *Worker) WithClock(now func() time.Time) *Worker {
	w.now = now
	return w
}

// Run memproses satu batch job yang sudah jatuh tempo.
func (w *Worker) Run(ctx context.Context) (Result, error) {
	res := Result{}
	w.loadConfig(ctx)
	now := w.now().UTC()
	claimed, err := w.jobs.ClaimDue(ctx, w.batchSize, now)
	if err != nil {
		return res, fmt.Errorf("claim router jobs: %w", err)
	}
	consecutiveErrors := 0
	for _, job := range claimed {
		res.Processed++
		handler, ok := w.handlers[job.JobType]
		var execErr error
		if !ok {
			execErr = fmt.Errorf("no handler for job type %s", job.JobType)
		} else {
			execErr = safeExec(ctx, handler, job)
		}
		if execErr == nil {
			consecutiveErrors = 0
			if err := w.jobs.MarkSucceeded(ctx, job.ID); err != nil {
				logger.WithComponent("RouterJobWorker").WithFields(map[string]any{"job_id": job.ID}).
					WithError(err).Warn("mark job succeeded failed")
			}
			res.Succeeded++
			continue
		}
		msg := truncateError(execErr)
		logger.WithComponent("RouterJobWorker").WithFields(map[string]any{
			"job_id": job.ID, "job_type": string(job.JobType), "attempts": job.Attempts,
		}).WithError(execErr).Warn("router job failed")
		if job.CanRetry() {
			next := now.Add(domainRouterJob.Backoff(job.Attempts))
			if err := w.jobs.MarkRetry(ctx, job.ID, next, msg); err != nil {
				logger.WithComponent("RouterJobWorker").WithFields(map[string]any{"job_id": job.ID}).
					WithError(err).Warn("mark job retry failed")
			}
			res.Retried++
		} else {
			if err := w.jobs.MarkDead(ctx, job.ID, msg); err != nil {
				logger.WithComponent("RouterJobWorker").WithFields(map[string]any{"job_id": job.ID}).
					WithError(err).Warn("mark job dead failed")
			}
			res.Dead++
		}
		consecutiveErrors++
		if consecutiveErrors >= w.errorBudget {
			res.StoppedEarly = true
			logger.WithComponent("RouterJobWorker").WithFields(map[string]any{
				"consecutive_errors": consecutiveErrors,
			}).Warn("error budget tercapai; batch dihentikan")
			break
		}
	}
	return res, nil
}

// EnqueueProvision menambahkan job provisioning untuk satu langganan.
// Idempotency key disegmentasi per menit agar klik ganda tidak menggandakan
// job, tetapi retry manual pada menit berikutnya tetap membuat job baru.
func (w *Worker) EnqueueProvision(ctx context.Context, tenantID, deviceID, subscriptionID, serviceType string) (domainRouterJob.Job, error) {
	jobType, err := JobTypeForService(serviceType)
	if err != nil {
		return domainRouterJob.Job{}, err
	}
	if strings.TrimSpace(deviceID) == "" {
		return domainRouterJob.Job{}, fmt.Errorf("device_id is required")
	}
	payload, _ := json.Marshal(map[string]string{"subscription_id": subscriptionID})
	now := w.now().UTC()
	job := domainRouterJob.Job{
		ID:             idgen.New("rjob"),
		TenantID:       orDefaultTenant(tenantID),
		DeviceID:       deviceID,
		JobType:        jobType,
		Payload:        payload,
		IdempotencyKey: fmt.Sprintf("provision:%s:%d", subscriptionID, now.Unix()/60),
		Status:         domainRouterJob.StatusQueued,
		MaxAttempts:    5,
		NextRunAt:      now,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	return w.jobs.Enqueue(ctx, job)
}

// EnqueueSyncProfile menambahkan job sinkronisasi profil paket ke router.
func (w *Worker) EnqueueSyncProfile(ctx context.Context, tenantID, deviceID, planID string) (domainRouterJob.Job, error) {
	if strings.TrimSpace(deviceID) == "" || strings.TrimSpace(planID) == "" {
		return domainRouterJob.Job{}, fmt.Errorf("device_id and plan_id are required")
	}
	payload, _ := json.Marshal(map[string]string{"plan_id": planID})
	now := w.now().UTC()
	job := domainRouterJob.Job{
		ID:             idgen.New("rjob"),
		TenantID:       orDefaultTenant(tenantID),
		DeviceID:       deviceID,
		JobType:        domainRouterJob.TypeSyncProfile,
		Payload:        payload,
		IdempotencyKey: fmt.Sprintf("sync-profile:%s:%d", planID, now.Unix()/60),
		Status:         domainRouterJob.StatusQueued,
		MaxAttempts:    3,
		NextRunAt:      now,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	return w.jobs.Enqueue(ctx, job)
}

// JobTypeForService memetakan service_type langganan ke tipe job provisi.
func JobTypeForService(serviceType string) (domainRouterJob.Type, error) {
	switch strings.ToLower(strings.TrimSpace(serviceType)) {
	case "", "pppoe":
		return domainRouterJob.TypeProvisionPPPoE, nil
	case "hotspot":
		return domainRouterJob.TypeProvisionHotspot, nil
	case "dedicated":
		return domainRouterJob.TypeProvisionDedicated, nil
	default:
		return "", fmt.Errorf("unknown service type %q", serviceType)
	}
}

// ConfigFromSettings membaca parameter worker dari system_settings.
func (w *Worker) loadConfig(ctx context.Context) {
	if w.settings == nil {
		return
	}
	if v := atoiSetting(w.settings.GetValue(ctx, "isp.router_job_batch_size", "20")); v > 0 {
		w.batchSize = v
	}
	if v := atoiSetting(w.settings.GetValue(ctx, "isp.router_job_error_budget", "5")); v > 0 {
		w.errorBudget = v
	}
}

func safeExec(ctx context.Context, h Handler, job domainRouterJob.Job) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("handler panic: %v", r)
		}
	}()
	return h(ctx, job)
}

func truncateError(err error) string {
	msg := err.Error()
	const maxLen = 500
	if len(msg) > maxLen {
		return msg[:maxLen]
	}
	return msg
}

func atoiSetting(s string) int {
	n := 0
	if _, err := fmt.Sscanf(strings.TrimSpace(s), "%d", &n); err != nil {
		return 0
	}
	return n
}

func orDefaultTenant(t string) string {
	if strings.TrimSpace(t) == "" {
		return domainTenant.DefaultID
	}
	return t
}

// BatchSize mengembalikan ukuran batch efektif (untuk observability/test).
func (w *Worker) BatchSize() int { return w.batchSize }

// ErrorBudget mengembalikan error budget efektif (untuk observability/test).
func (w *Worker) ErrorBudget() int { return w.errorBudget }
