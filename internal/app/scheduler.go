package app

import (
	"context"
	"time"

	cronv3 "github.com/robfig/cron/v3"

	billingUC "github.com/quixiq/polyglot/internal/usecase/billing"
	"github.com/quixiq/polyglot/pkg/logger"
)

// Scheduler menjalankan pekerjaan periodik ISP (fase 3):
//   - generator tagihan bulanan (BILLING_CRON, default 06:00 harian)
//   - worker isolir otomatis     (ISOLATION_CRON, default tiap 10 menit)
//
// Tiap job dilindungi recover + logging; scheduler berhenti via ctx cancel
// bersama graceful shutdown aplikasi.
type Scheduler struct {
	cron      *cronv3.Cron
	billingUC *billingUC.RunBillingUseCase
	isolateUC *billingUC.IsolateWorker
	tenantID  string
}

func newScheduler(billing *billingUC.RunBillingUseCase, isolate *billingUC.IsolateWorker, billingSpec, isolationSpec, tenantID string) *Scheduler {
	s := &Scheduler{
		cron:      cronv3.New(cronv3.WithLocation(time.Local)),
		billingUC: billing,
		isolateUC: isolate,
		tenantID:  tenantID,
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
				Info("scheduler job finished")
		}
	}

	s.cron.AddFunc(billingSpec, safeWrap("run-billing", func(ctx context.Context) error {
		res, err := s.billingUC.Run(ctx, s.tenantID, time.Now().Format("2006-01"))
		if err == nil {
			logger.WithComponent("Scheduler").WithFields(map[string]any{
				"created": res.Created, "skipped": res.Skipped,
			}).Info("billing run summary")
		}
		return err
	}))
	s.cron.AddFunc(isolationSpec, safeWrap("isolate-worker", func(ctx context.Context) error {
		res, err := s.isolateUC.Run(ctx)
		if err == nil && res.Isolated+res.RouterFailures+res.Provisioned+res.Suspended > 0 {
			logger.WithComponent("Scheduler").WithFields(map[string]any{
				"isolated": res.Isolated, "suspended": res.Suspended,
				"provisioned": res.Provisioned, "provision_failed": res.ProvisionFailed,
				"router_failures":   res.RouterFailures,
				"skipped_no_router": res.SkippedNoRouter,
			}).Info("lifecycle pass summary")
		}
		return err
	}))
	return s
}

func (s *Scheduler) Start() { s.cron.Start() }

func (s *Scheduler) Stop() { s.cron.Stop() }
