package report

import (
	"context"
	"testing"
	"time"

	domainReporting "github.com/quixiq/polyglot/internal/domain/reporting"
)

type mockReportRepo struct {
	snaps []domainReporting.DailyFinancialSnapshot
}

func (m *mockReportRepo) GetByDate(ctx context.Context, tenantID string, date time.Time) (domainReporting.DailyFinancialSnapshot, error) {
	return domainReporting.DailyFinancialSnapshot{
		SnapshotDate: date,
		InvoiceCount: 5,
		InvoiceTotal: 500000,
	}, nil
}

func (m *mockReportRepo) ListRange(ctx context.Context, tenantID string, from, to time.Time) ([]domainReporting.DailyFinancialSnapshot, error) {
	return m.snaps, nil
}

func (m *mockReportRepo) UpsertSnapshot(ctx context.Context, snap domainReporting.DailyFinancialSnapshot) error {
	m.snaps = append(m.snaps, snap)
	return nil
}

type mockSnapshotComputer struct {
	recomputed bool
}

func (m *mockSnapshotComputer) RecomputeDaily(ctx context.Context, tenantID string, date time.Time) error {
	m.recomputed = true
	return nil
}

func TestManageReportUseCase(t *testing.T) {
	ctx := context.Background()
	repo := &mockReportRepo{
		snaps: []domainReporting.DailyFinancialSnapshot{
			{SnapshotDate: time.Now(), InvoiceCount: 2, InvoiceTotal: 200000},
		},
	}
	comp := &mockSnapshotComputer{}
	uc := NewManageReportUseCase(repo, comp)

	// Test DailyReport
	period, dailySnaps, err := uc.DailyReport(ctx, "2026-09-01")
	if err != nil || period != "2026-09-01" || len(dailySnaps) != 1 {
		t.Fatalf("unexpected daily report result: %s, %v, %v", period, dailySnaps, err)
	}

	// Test MonthlyReport
	mPeriod, mSnaps, err := uc.MonthlyReport(ctx, "2026-09")
	if err != nil || mPeriod != "2026-09" || len(mSnaps) != 1 {
		t.Fatalf("unexpected monthly report result: %s, %v, %v", mPeriod, mSnaps, err)
	}

	// Test YearlyReport
	yPeriod, ySnaps, err := uc.YearlyReport(ctx, 2026)
	if err != nil || yPeriod != "2026" || len(ySnaps) != 1 {
		t.Fatalf("unexpected yearly report result: %s, %v, %v", yPeriod, ySnaps, err)
	}

	// Test RefreshSnapshot
	refDate, err := uc.RefreshSnapshot(ctx, "2026-09-01")
	if err != nil || refDate != "2026-09-01" || !comp.recomputed {
		t.Fatalf("unexpected refresh snapshot result: %s, %v", refDate, err)
	}
}
