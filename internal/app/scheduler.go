package app

import (
	"context"
	"fmt"
	"time"

	cronv3 "github.com/robfig/cron/v3"

	"github.com/quixiq/polyglot/internal/port"
	billingUC "github.com/quixiq/polyglot/internal/usecase/billing"
	notificationUC "github.com/quixiq/polyglot/internal/usecase/notification"
	"github.com/quixiq/polyglot/pkg/logger"
)

// schedulerJobs menghimpun seluruh pekerjaan periodik ISP.
type schedulerJobs struct {
	billing        *billingUC.RunBillingUseCase
	isolate        *billingUC.IsolateWorker
	reminder       *billingUC.ReminderWorker
	waSend         *notificationUC.WASenderWorker
	routerJobs     func(ctx context.Context) error
	drift          func(ctx context.Context) error
	snapshot       func(ctx context.Context) error
	oltSync        func(ctx context.Context) error
	acsMatch       func(ctx context.Context) error
	acsTelemetry   func(ctx context.Context) error
	approvalExpire func(ctx context.Context) error
	restoreCleanup func(ctx context.Context) error
	backupDaily    func(ctx context.Context) error
	backupWeekly   func(ctx context.Context) error
	retention      func(ctx context.Context) error
}

type schedulerSpecs struct {
	billing, isolation, reminder, waSend, snapshot, routerJobs, drift, oltSync, acsMatch, acsTelemetry string
	approvalExpire, restoreCleanup, backupDaily, backupWeekly, retention                               string
}

// Scheduler menjalankan pekerjaan periodik ISP:
//   - generator tagihan bulanan (BILLING_CRON, default 06:00 harian)
//   - lifecycle worker          (ISOLATION_CRON, default tiap 10 menit)
//   - pengirim WA               (WA_SEND_CRON, default tiap 30 detik)
//   - snapshot harian           (SNAPSHOT_CRON, default 00:05 harian)
//   - worker antrean job router (ROUTER_JOB_CRON, default tiap 1 menit)
//   - drift reconciliation      (DRIFT_CRON, default 03:00 harian)
//
// Tiap job dilindungi recover + logging; berhenti bersama graceful shutdown.
type Scheduler struct {
	cron     *cronv3.Cron
	jobs     schedulerJobs
	tenantID string
	locker   port.JobLocker
}

func newScheduler(jobs schedulerJobs, specs schedulerSpecs, tenantID string, locker port.JobLocker) *Scheduler {
	s := &Scheduler{
		// Cron mengikuti timezone server (atur via env TZ); seluruh timestamp
		// data yang dihitung job memakai UTC (F6-9).
		cron:     cronv3.New(cronv3.WithLocation(time.Local)),
		jobs:     jobs,
		tenantID: tenantID,
		locker:   locker,
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
		locked := func(ctx context.Context) error {
			if s.locker != nil {
				unlock, ok, lerr := s.locker.TryLock(ctx, name)
				if lerr != nil {
					logger.WithComponent("Scheduler").WithFields(map[string]any{"job": name}).
						WithError(lerr).Warn("gagal mengambil job lock; job dilewati siklus ini")
					return nil
				}
				if !ok {
					logger.WithComponent("Scheduler").WithFields(map[string]any{"job": name}).
						Debug("job dilewati: instance lain memegang lock")
					return nil
				}
				defer unlock()
			}
			return fn(ctx)
		}
		if _, err := s.cron.AddFunc(spec, safeWrap(name, locked)); err != nil {
			logger.WithComponent("Scheduler").WithFields(map[string]any{
				"job": name, "spec": spec,
			}).WithError(err).Error("spec cron tidak valid — job dilewati")
		}
	}

	if s.jobs.billing != nil {
		add(specs.billing, "run-billing", func(ctx context.Context) error {
			// Jalankan periode berjalan + periode depan: mode PRABAYAR
			// menerbitkan invoice lebih awal (lead days), sedangkan
			// PASCABAYAR otomatis melewati periode depan karena belum due.
			now := time.Now().UTC()
			periods := []string{now.Format("2006-01"), now.AddDate(0, 1, 0).Format("2006-01")}
			totalCreated, totalSkipped := 0, 0
			for _, period := range periods {
				res, err := s.jobs.billing.Run(ctx, s.tenantID, period)
				if err != nil {
					return err
				}
				totalCreated += res.Created
				totalSkipped += res.Skipped
			}
			logger.WithComponent("Scheduler").WithFields(map[string]any{
				"created": totalCreated, "skipped": totalSkipped,
			}).Info("billing run summary")
			return nil
		})
	}
	if s.jobs.isolate != nil {
		add(specs.isolation, "lifecycle-worker", func(ctx context.Context) error {
			res, err := s.jobs.isolate.Run(ctx)
			if err == nil && res.Isolated+res.RouterFailures+res.Provisioned+res.Suspended+res.Errors > 0 {
				logger.WithComponent("Scheduler").WithFields(map[string]any{
					"isolated": res.Isolated, "suspended": res.Suspended,
					"provisioned": res.Provisioned, "provision_failed": res.ProvisionFailed,
					"router_failures":   res.RouterFailures,
					"skipped_no_router": res.SkippedNoRouter,
					"errors":            res.Errors,
				}).Info("lifecycle pass summary")
			}
			return err
		})
	}
	if s.jobs.routerJobs != nil {
		add(specs.routerJobs, "router-job-worker", func(ctx context.Context) error {
			return s.jobs.routerJobs(ctx)
		})
	}
	if s.jobs.drift != nil {
		add(specs.drift, "drift-reconciliation", func(ctx context.Context) error {
			return s.jobs.drift(ctx)
		})
	}
	if s.jobs.reminder != nil {
		add(specs.reminder, "bill-reminder", func(ctx context.Context) error {
			res, err := s.jobs.reminder.Run(ctx)
			if err != nil {
				return fmt.Errorf("bill reminder: %w", err)
			}
			if res.Queued+res.Skipped > 0 {
				logger.WithComponent("Scheduler").WithFields(map[string]any{
					"queued": res.Queued, "skipped": res.Skipped,
				}).Info("bill reminder summary")
			}
			return nil
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
	if s.jobs.approvalExpire != nil {
		add(specs.approvalExpire, "approval-expire", func(ctx context.Context) error {
			return s.jobs.approvalExpire(ctx)
		})
	}
	if s.jobs.restoreCleanup != nil {
		add(specs.restoreCleanup, "restore-cleanup", func(ctx context.Context) error {
			return s.jobs.restoreCleanup(ctx)
		})
	}
	if s.jobs.backupDaily != nil {
		add(specs.backupDaily, "backup-daily", func(ctx context.Context) error {
			return s.jobs.backupDaily(ctx)
		})
	}
	if s.jobs.retention != nil {
		add(specs.retention, "ai-retention", func(ctx context.Context) error {
			return s.jobs.retention(ctx)
		})
	}
	if s.jobs.backupWeekly != nil {
		add(specs.backupWeekly, "backup-weekly", func(ctx context.Context) error {
			return s.jobs.backupWeekly(ctx)
		})
	}
	if s.jobs.oltSync != nil {
		add(specs.oltSync, "olt-sync", func(ctx context.Context) error {
			return s.jobs.oltSync(ctx)
		})
	}
	if s.jobs.acsMatch != nil {
		add(specs.acsMatch, "acs-match", func(ctx context.Context) error {
			return s.jobs.acsMatch(ctx)
		})
	}
	if s.jobs.acsTelemetry != nil {
		add(specs.acsTelemetry, "acs-telemetry", func(ctx context.Context) error {
			return s.jobs.acsTelemetry(ctx)
		})
	}
	return s
}

func (s *Scheduler) Start() { s.cron.Start() }

func (s *Scheduler) Stop() { s.cron.Stop() }
