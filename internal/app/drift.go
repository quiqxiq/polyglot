package app

import (
	"context"
	"strings"

	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/internal/usecase/importer"
	"github.com/quixiq/polyglot/pkg/logger"
)

// runDriftReconciliation membandingkan akun langganan aktif di DB dengan
// secrets PPP di setiap router MikroTik. Hanya menghasilkan laporan drift
// (missing in DB / missing in router / profile mismatch) — tidak pernah
// menghapus akun sepihak (planning v2 Fase 4).
func runDriftReconciliation(
	ctx context.Context,
	devices port.DeviceRepository,
	resolver func(context.Context, string) (port.DeviceDriver, bool),
	rec *importer.Reconciler,
) error {
	if devices == nil || rec == nil {
		return nil
	}
	list, err := devices.FindAll(ctx)
	if err != nil {
		return err
	}
	for _, dev := range list {
		if !dev.Enabled {
			continue
		}
		if !strings.Contains(strings.ToLower(dev.DriverType), "mikrotik") {
			continue
		}
		driver, ok := resolver(ctx, dev.ID)
		if !ok {
			logger.WithComponent("DriftReconciliation").WithFields(map[string]any{
				"device_id": dev.ID,
			}).Warn("driver resolution failed; device skipped")
			continue
		}
		report, cerr := rec.Compare(ctx, dev.ID, driver)
		if cerr != nil {
			logger.WithComponent("DriftReconciliation").WithFields(map[string]any{
				"device_id": dev.ID,
			}).WithError(cerr).Warn("drift check failed")
			continue
		}
		driftCount := len(report.MissingInDB) + len(report.MissingInRoute) + len(report.ProfileMismatch)
		if driftCount == 0 {
			logger.WithComponent("DriftReconciliation").WithFields(map[string]any{
				"device_id": dev.ID, "device": dev.Name,
			}).Info("no drift detected")
			continue
		}
		logger.WithComponent("DriftReconciliation").WithFields(map[string]any{
			"device_id":               dev.ID,
			"device":                  dev.Name,
			"missing_in_db":           len(report.MissingInDB),
			"missing_in_router":       len(report.MissingInRoute),
			"profile_mismatch":        len(report.ProfileMismatch),
			"sample_missing_in_db":    sample(report.MissingInDB),
			"sample_missing_in_route": sample(report.MissingInRoute),
		}).Warn("router drift detected (report only, no deletion)")
	}
	return nil
}

func sample(items []string) []string {
	const max = 5
	if len(items) <= max {
		return items
	}
	return items[:max]
}
