package app

import (
	"context"
	"fmt"
	"time"

	"github.com/quixiq/polyglot/internal/adapter/postgres"
	"github.com/quixiq/polyglot/internal/adapter/provisioner"
	waAdapter "github.com/quixiq/polyglot/internal/adapter/whatsapp"
	"github.com/quixiq/polyglot/internal/config"
	domainTenant "github.com/quixiq/polyglot/internal/domain/tenant"
	"github.com/quixiq/polyglot/internal/port"
	billingUC "github.com/quixiq/polyglot/internal/usecase/billing"
	"github.com/quixiq/polyglot/internal/usecase/importer"
	notificationUC "github.com/quixiq/polyglot/internal/usecase/notification"
	oltUC "github.com/quixiq/polyglot/internal/usecase/olt"
	ontUC "github.com/quixiq/polyglot/internal/usecase/ont"
	jobUCRouter "github.com/quixiq/polyglot/internal/usecase/routerjob"
	"github.com/quixiq/polyglot/pkg/logger"
)

// schedulerDeps groups the dependencies needed to build the periodic ISP
// workers. Keeping them explicit avoids a second hidden container inside New.
type schedulerDeps struct {
	cfg            config.Config
	pgStore        *postgres.Store
	subscriptions  *postgres.SubscriptionRepository
	invoices       *postgres.InvoiceRepository
	customers      *postgres.CustomerRepository
	plans          *postgres.ServicePlanRepository
	accountManager *provisioner.Provisioner
	notifications  *postgres.NotificationRepository
	settings       *postgres.SettingRepository
	waSender       *waAdapter.SenderAdapter
	reporting      *postgres.ReportingRepository
	routerJobs     *jobUCRouter.Worker
	devices        port.DeviceRepository
	driverResolver func(context.Context, string) (port.DeviceDriver, bool)
	reconciler     *importer.Reconciler
	ontService     *ontUC.Service
	oltService     *oltUC.Service
	billing        *billingUC.RunBillingUseCase
}

// buildScheduler constructs and starts the ISP cron scheduler. It returns nil
// when the scheduler is disabled or when the mandatory cron specs are absent.
func buildScheduler(ctx context.Context, d schedulerDeps) *Scheduler {
	if !d.cfg.SchedulerEnabled || d.cfg.BillingCronSpec == "" || d.cfg.IsolationCronSpec == "" {
		return nil
	}

	isolateWorker := billingUC.NewIsolateWorker(
		d.subscriptions, d.invoices, d.customers, d.plans, d.accountManager, d.notifications, d.settings,
	)
	reminderWorker := billingUC.NewReminderWorker(d.subscriptions, d.invoices, d.customers, d.notifications).
		WithSettings(d.settings)
	waWorker := notificationUC.NewWASenderWorker(d.notifications, d.waSender, d.settings)
	snapshotJob := func(ctx context.Context) error {
		return d.reporting.RecomputeDaily(ctx, domainTenant.DefaultID, timeNowUTC())
	}
	routerJobsJob := func(ctx context.Context) error {
		res, rerr := d.routerJobs.Run(ctx)
		if res.Processed > 0 {
			logger.WithComponent("App").WithFields(map[string]any{
				"processed": res.Processed, "succeeded": res.Succeeded,
				"retried": res.Retried, "dead": res.Dead,
				"stopped_early": res.StoppedEarly,
			}).Info("router job batch summary")
		}
		if rerr != nil {
			return fmt.Errorf("router job batch: %w", rerr)
		}
		return nil
	}
	driftJob := func(ctx context.Context) error {
		if err := runDriftReconciliation(ctx, d.devices, d.driverResolver, d.reconciler); err != nil {
			return fmt.Errorf("drift reconciliation: %w", err)
		}
		return nil
	}
	acsMatchJob := func(ctx context.Context) error {
		stats, aerr := d.ontService.SyncDevices(ctx)
		if aerr == nil {
			logger.WithComponent("App").WithFields(map[string]any{
				"devices": stats.Devices, "matched": stats.Matched,
				"unmatched": stats.Unmatched, "conflicts": stats.Conflicts,
			}).Info("acs match summary")
		} else {
			return fmt.Errorf("acs sync devices: %w", aerr)
		}
		return nil
	}
	acsTelemetryJob := func(ctx context.Context) error {
		ok, terr := d.ontService.RefreshTelemetryAll(ctx)
		if terr == nil {
			logger.WithComponent("App").WithField("updated", ok).Info("acs telemetry summary")
		} else {
			return fmt.Errorf("acs telemetry refresh: %w", terr)
		}
		return nil
	}
	oltSyncJob := func(ctx context.Context) error {
		res, serr := d.oltService.SyncAll(ctx)
		for deviceID, r := range res {
			logger.WithComponent("App").WithFields(map[string]any{
				"device_id": deviceID, "discovered": r.Discovered, "updated": r.Updated,
				"removed": r.Removed, "online": r.Online, "offline": r.Offline,
			}).Info("olt sync summary")
		}
		if serr != nil {
			return fmt.Errorf("olt sync all: %w", serr)
		}
		return nil
	}

	sched := newScheduler(schedulerJobs{
		billing:      d.billing,
		isolate:      isolateWorker,
		reminder:     reminderWorker,
		waSend:       waWorker,
		routerJobs:   routerJobsJob,
		drift:        driftJob,
		snapshot:     snapshotJob,
		oltSync:      oltSyncJob,
		acsMatch:     acsMatchJob,
		acsTelemetry: acsTelemetryJob,
	}, schedulerSpecs{
		billing:      d.cfg.BillingCronSpec,
		isolation:    d.cfg.IsolationCronSpec,
		reminder:     d.cfg.ReminderCronSpec,
		waSend:       d.cfg.WaSendCronSpec,
		routerJobs:   d.cfg.RouterJobCronSpec,
		drift:        d.cfg.DriftCronSpec,
		snapshot:     d.cfg.SnapshotCronSpec,
		oltSync:      d.cfg.OltSyncCronSpec,
		acsMatch:     d.cfg.ACSMatchCronSpec,
		acsTelemetry: d.cfg.ACSTelemetryCron,
	}, domainTenant.DefaultID, postgres.NewAdvisoryLocker(d.pgStore.DB()))
	sched.Start()
	logger.WithComponent("App").WithFields(map[string]any{
		"billing_cron":    d.cfg.BillingCronSpec,
		"isolation_cron":  d.cfg.IsolationCronSpec,
		"reminder_cron":   d.cfg.ReminderCronSpec,
		"wa_send_cron":    d.cfg.WaSendCronSpec,
		"snapshot_cron":   d.cfg.SnapshotCronSpec,
		"router_job_cron": d.cfg.RouterJobCronSpec,
		"drift_cron":      d.cfg.DriftCronSpec,
	}).Info("ISP scheduler started")
	return sched
}

func timeNowUTC() time.Time {
	n := time.Now().UTC()
	return time.Date(n.Year(), n.Month(), n.Day(), 0, 0, 0, 0, time.UTC)
}
