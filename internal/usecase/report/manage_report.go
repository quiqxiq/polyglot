package report

import (
	"context"
	"fmt"
	"strconv"
	"time"

	domainReporting "github.com/quixiq/polyglot/internal/domain/reporting"
	"github.com/quixiq/polyglot/internal/port"
)

// ManageReportUseCase orchestrates financial and operational reporting snapshots.
type ManageReportUseCase struct {
	repo        port.ReportingRepository
	snapshotter port.SnapshotComputer
}

// NewManageReportUseCase constructs a new ManageReportUseCase.
func NewManageReportUseCase(repo port.ReportingRepository, snapshotter port.SnapshotComputer) *ManageReportUseCase {
	return &ManageReportUseCase{
		repo:        repo,
		snapshotter: snapshotter,
	}
}

// DailyReport retrieves the daily financial snapshot for a given date.
func (uc *ManageReportUseCase) DailyReport(ctx context.Context, dateStr string) (string, []domainReporting.DailyFinancialSnapshot, error) {
	if uc.repo == nil {
		return "", nil, domainReporting.ErrNotFound
	}
	day := time.Now().UTC()
	if dateStr != "" {
		d, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			return "", nil, fmt.Errorf("%w: date must be YYYY-MM-DD", domainReporting.ErrInvalidInput)
		}
		day = d
	}
	snap, err := uc.repo.GetByDate(ctx, "tenant-default", day)
	if err != nil {
		return "", nil, fmt.Errorf("get daily report: %w", err)
	}
	return day.Format("2006-01-02"), []domainReporting.DailyFinancialSnapshot{snap}, nil
}

// MonthlyReport retrieves daily financial snapshots across a given month (YYYY-MM).
func (uc *ManageReportUseCase) MonthlyReport(ctx context.Context, monthStr string) (string, []domainReporting.DailyFinancialSnapshot, error) {
	if uc.repo == nil {
		return "", nil, domainReporting.ErrNotFound
	}
	month := monthStr
	if month == "" {
		month = time.Now().Format("2006-01")
	}
	from, err := time.Parse("2006-01", month)
	if err != nil {
		return "", nil, fmt.Errorf("%w: month must be YYYY-MM", domainReporting.ErrInvalidInput)
	}
	to := from.AddDate(0, 1, -1)
	snaps, err := uc.repo.ListRange(ctx, "tenant-default", from, to)
	if err != nil {
		return "", nil, fmt.Errorf("list monthly range: %w", err)
	}
	return month, snaps, nil
}

// YearlyReport retrieves daily financial snapshots across a given year.
func (uc *ManageReportUseCase) YearlyReport(ctx context.Context, year int) (string, []domainReporting.DailyFinancialSnapshot, error) {
	if uc.repo == nil {
		return "", nil, domainReporting.ErrNotFound
	}
	if year == 0 {
		year = time.Now().Year()
	}
	from := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	to := from.AddDate(1, 0, -1)
	snaps, err := uc.repo.ListRange(ctx, "tenant-default", from, to)
	if err != nil {
		return "", nil, fmt.Errorf("list yearly range: %w", err)
	}
	return strconv.Itoa(year), snaps, nil
}

// RefreshSnapshot recomputes the daily snapshot for the specified date.
func (uc *ManageReportUseCase) RefreshSnapshot(ctx context.Context, dateStr string) (string, error) {
	if uc.snapshotter == nil {
		return "", domainReporting.ErrNotFound
	}
	day := time.Now().UTC()
	if dateStr != "" {
		d, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			return "", fmt.Errorf("%w: date must be YYYY-MM-DD", domainReporting.ErrInvalidInput)
		}
		day = d
	}
	if err := uc.snapshotter.RecomputeDaily(ctx, "tenant-default", day); err != nil {
		return "", fmt.Errorf("recompute daily snapshot: %w", err)
	}
	return day.Format("2006-01-02"), nil
}
