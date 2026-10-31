package report

import (
	"context"
	"fmt"
	"strings"
	"time"

	domainReporting "github.com/quixiq/polyglot/internal/domain/reporting"
	"github.com/quixiq/polyglot/internal/port"
)

// WithAdvancedRepository menautkan laporan operasional lanjutan (AR Aging,
// rekonsiliasi gateway, statistik PSB) — planning v2 Fase 5.
func (uc *ManageReportUseCase) WithAdvancedRepository(r port.AdvancedReportRepository) *ManageReportUseCase {
	uc.advanced = r
	return uc
}

// ARAging menghitung umur piutang per bucket pada tanggal tertentu.
func (uc *ManageReportUseCase) ARAging(ctx context.Context, asOf string) (domainReporting.ARAgingReport, error) {
	if uc.advanced == nil {
		return domainReporting.ARAgingReport{}, fmt.Errorf("%w: advanced reporting unavailable", domainReporting.ErrNotFound)
	}
	day := time.Now().UTC()
	if strings.TrimSpace(asOf) != "" {
		d, err := time.Parse("2006-01-02", strings.TrimSpace(asOf))
		if err != nil {
			return domainReporting.ARAgingReport{}, fmt.Errorf("%w: as_of must be YYYY-MM-DD", domainReporting.ErrInvalidInput)
		}
		day = d
	}
	return uc.advanced.ARAging(ctx, "tenant-default", day)
}

// GatewayReconciliation membandingkan transaksi gateway dengan kas tercatat.
func (uc *ManageReportUseCase) GatewayReconciliation(ctx context.Context, fromStr, toStr string) (domainReporting.GatewayReconciliationReport, error) {
	if uc.advanced == nil {
		return domainReporting.GatewayReconciliationReport{}, fmt.Errorf("%w: advanced reporting unavailable", domainReporting.ErrNotFound)
	}
	from, to, err := parseRange(fromStr, toStr)
	if err != nil {
		return domainReporting.GatewayReconciliationReport{}, err
	}
	return uc.advanced.GatewayReconciliation(ctx, "tenant-default", from, to)
}

// PSBStats menghitung efisiensi & durasi penyelesaian PSB.
func (uc *ManageReportUseCase) PSBStats(ctx context.Context, fromStr, toStr string) (domainReporting.PSBStatsReport, error) {
	if uc.advanced == nil {
		return domainReporting.PSBStatsReport{}, fmt.Errorf("%w: advanced reporting unavailable", domainReporting.ErrNotFound)
	}
	from, to, err := parseRange(fromStr, toStr)
	if err != nil {
		return domainReporting.PSBStatsReport{}, err
	}
	return uc.advanced.PSBStats(ctx, "tenant-default", from, to)
}

// parseRange mem-parse rentang tanggal; default: awal bulan berjalan s/d besok.
func parseRange(fromStr, toStr string) (time.Time, time.Time, error) {
	now := time.Now().UTC()
	from := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC).AddDate(0, 0, 1)
	if strings.TrimSpace(fromStr) != "" {
		d, err := time.Parse("2006-01-02", strings.TrimSpace(fromStr))
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("%w: from must be YYYY-MM-DD", domainReporting.ErrInvalidInput)
		}
		from = d
	}
	if strings.TrimSpace(toStr) != "" {
		d, err := time.Parse("2006-01-02", strings.TrimSpace(toStr))
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("%w: to must be YYYY-MM-DD", domainReporting.ErrInvalidInput)
		}
		to = d.AddDate(0, 0, 1)
	}
	return from, to, nil
}

// TaxReport menghitung rekap PPN faktur pada periode YYYY-MM.
func (uc *ManageReportUseCase) TaxReport(ctx context.Context, period string) (domainReporting.TaxReport, error) {
	if uc.advanced == nil {
		return domainReporting.TaxReport{}, fmt.Errorf("%w: advanced reporting unavailable", domainReporting.ErrNotFound)
	}
	if strings.TrimSpace(period) == "" {
		period = time.Now().UTC().Format("2006-01")
	}
	return uc.advanced.TaxReport(ctx, "tenant-default", period)
}

// SalesReport menghitung pendapatan penjualan per paket.
func (uc *ManageReportUseCase) SalesReport(ctx context.Context, period string) (domainReporting.SalesReport, error) {
	if uc.advanced == nil {
		return domainReporting.SalesReport{}, fmt.Errorf("%w: advanced reporting unavailable", domainReporting.ErrNotFound)
	}
	if strings.TrimSpace(period) == "" {
		period = time.Now().UTC().Format("2006-01")
	}
	return uc.advanced.SalesReport(ctx, "tenant-default", period)
}

// CollectionReport menghitung efektivitas penagihan pada periode YYYY-MM.
func (uc *ManageReportUseCase) CollectionReport(ctx context.Context, period string) (domainReporting.CollectionReport, error) {
	if uc.advanced == nil {
		return domainReporting.CollectionReport{}, fmt.Errorf("%w: advanced reporting unavailable", domainReporting.ErrNotFound)
	}
	if strings.TrimSpace(period) == "" {
		period = time.Now().UTC().Format("2006-01")
	}
	return uc.advanced.CollectionReport(ctx, "tenant-default", period)
}

// CashflowReport merangkum mutasi kas masuk dan keluar per kategori.
func (uc *ManageReportUseCase) CashflowReport(ctx context.Context, fromStr, toStr string) (domainReporting.CashflowReport, error) {
	if uc.advanced == nil {
		return domainReporting.CashflowReport{}, fmt.Errorf("%w: advanced reporting unavailable", domainReporting.ErrNotFound)
	}
	from, to, err := parseRange(fromStr, toStr)
	if err != nil {
		return domainReporting.CashflowReport{}, err
	}
	return uc.advanced.CashflowReport(ctx, "tenant-default", from, to)
}

// AgingDetail menampilkan rincian piutang per pelanggan.
func (uc *ManageReportUseCase) AgingDetail(ctx context.Context, asOf string, limit int) (domainReporting.AgingDetailReport, error) {
	if uc.advanced == nil {
		return domainReporting.AgingDetailReport{}, fmt.Errorf("%w: advanced reporting unavailable", domainReporting.ErrNotFound)
	}
	day := time.Now().UTC()
	if strings.TrimSpace(asOf) != "" {
		d, err := time.Parse("2006-01-02", strings.TrimSpace(asOf))
		if err != nil {
			return domainReporting.AgingDetailReport{}, fmt.Errorf("%w: as_of must be YYYY-MM-DD", domainReporting.ErrInvalidInput)
		}
		day = d
	}
	return uc.advanced.AgingDetail(ctx, "tenant-default", day, limit)
}

// ClosingReport merangkum performa dan saldo pada tanggal tutup buku.
func (uc *ManageReportUseCase) ClosingReport(ctx context.Context, period string) (domainReporting.ClosingReport, error) {
	if uc.advanced == nil {
		return domainReporting.ClosingReport{}, fmt.Errorf("%w: advanced reporting unavailable", domainReporting.ErrNotFound)
	}
	if strings.TrimSpace(period) == "" {
		period = time.Now().UTC().Format("2006-01")
	}
	closingDate := time.Now().UTC()
	if p, err := time.Parse("2006-01", period); err == nil {
		closingDate = time.Date(p.Year(), p.Month()+1, 0, 0, 0, 0, 0, time.UTC)
	}
	return uc.advanced.ClosingReport(ctx, "tenant-default", period, closingDate)
}
