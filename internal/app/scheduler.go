package app

import (
	"context"
	"time"

	cronv3 "github.com/robfig/cron/v3"

	billingUC "github.com/quixiq/polyglot/internal/usecase/billing"
	notificationUC "github.com/quixiq/polyglot/internal/usecase/notification"
	"github.com/quixiq/polyglot/pkg/logger"
)

// schedulerJobs menghimpun seluruh pekerjaan periodik ISP.
type schedulerJobs struct {
	billing  *billingUC.RunBillingUseCase
	isolate  *billingUC.IsolateWorker
	waSend   *notificationUC.WaSenderWorker
	snapshot func(ctx context.Context) error
}

type schedulerSpecs struct {
	billing, isolation, waSend, snapshot string
}

// Scheduler menjalankan pekerjaan periodik ISP:
//   - generator tagihan bulanan (BILLING_CRON, default 06:00 harian)
//   - lifecycle worker          (ISOLATION_CRON, default tiap 10 menit)
//   - pengirim WA               (WA_SEND_CRON, default tiap 30 detik)
//   - snapshot harian           (SNAPSHOT_CRON, default 00:05 harian)
//
// Tiap job dilindungi recover + logging; berhenti bersama graceful shutdown.
type Scheduler struct {
	cron     *cronv3.Cron
	jobs     schedulerJobs
	tenantID string
}

func newScheduler(jobs schedulerJobs, specs schedulerSpecs, tenantID string) *Scheduler {
	s := &Scheduler{
		cron:     cronv3.New(cronv3.WithLocation(time.Local)),
		jobs:     jobs,
		tenantID: tenantID,
	}

	safeWrap := func(name string, fn func(ctx context.Context) error) func() {
		return func() {
			defer func() {
				if r := recover(); r != nil {
					logger.WithComponent("Scheduler").WithFields(map[string]any{"job": name, "panic": r}).
						Error("scheduler job panicked")
				}
			}()
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
			defer cancel()
			if err := fn(ctx); err != nil {
				logger.WithComponent("Scheduler").WithFields(map[string]any{"job": name}).
					WithError(err).Error("scheduler job failed")
				return
			}
			logger.WithComponent("Scheduler").WithFields(map[string]any{"job": name}).
				Debug("scheduler job finished")
		}
	}

	add := func(spec, name string, fn func(ctx context.Context) error) {
		if spec == "" {
			return
		}
		if _, err := s.cron.AddFunc(spec, safeWrap(name, fn)); err != nil {
			logger.WithComponent("Scheduler").WithFields(map[string]any{
				"job": name, "spec": spec,
			}).WithError(err).Error("spec cron tidak valid — job dilewati")
		}
	}

	if s.jobs.billing != nil {
		add(specs.billing, "run-billing", func(ctx context.Context) error {
			res, err := s.jobs.billing.Run(ctx, s.tenantID, time.Now().Format("2006-01"))
			if err == nil {
				logger.WithComponent("Scheduler").WithFields(map[string]any{
					"created": res.Created, "skipped": res.Skipped,
				}).Info("billing run summary")
			}
			return err
		})
	}
	if s.jobs.isolate != nil {
		add(specs.isolation, "lifecycle-worker", func(ctx context.Context) error {
			res, err := s.jobs.isolate.Run(ctx)
			if err == nil && res.Isolated+res.RouterFailures+res.Provisioned+res.Suspended > 0 {
				logger.WithComponent("Scheduler").WithFields(map[string]any{
					"isolated": res.Isolated, "suspended": res.Suspended,
					"provisioned": res.Provisioned, "provision_failed": res.ProvisionFailed,
					"router_failures":   res.RouterFailures,
					"skipped_no_router": res.SkippedNoRouter,
				}).Info("lifecycle pass summary")
			}
			return err
		})
	}
	if s.jobs.waSend != nil {
		add(specs.waSend, "wa-sender", func(ctx context.Context) error {
			res, err := s.jobs.waSend.Run(ctx)
			if err == nil && res.Sent+res.Failed > 0 {
				logger.WithComponent("Scheduler").WithFields(map[string]any{
					"sent": res.Sent, "failed": res.Failed, "gave_up": res.GaveUp,
				}).Info("wa sender summary")
			}
			return err
		})
	}
	if s.jobs.snapshot != nil {
		add(specs.snapshot, "daily-snapshot", func(ctx context.Context) error {
			return s.jobs.snapshot(ctx)
		})
	}
	return s
}

func (s *Scheduler) Start() { s.cron.Start() }

func (s *Scheduler) Stop() { s.cron.Stop() }
