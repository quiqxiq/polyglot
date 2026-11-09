package report_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domainReporting "github.com/quixiq/polyglot/internal/domain/reporting"
	uc "github.com/quixiq/polyglot/internal/usecase/report"
)

type fakeAdvancedRepo struct {
	tenant string
	asOf   time.Time
	from   time.Time
	to     time.Time
}

func (f *fakeAdvancedRepo) ARAging(_ context.Context, tenant string, asOf time.Time) (domainReporting.ARAgingReport, error) {
	f.tenant, f.asOf = tenant, asOf
	return domainReporting.ARAgingReport{
		AsOf:       asOf,
		Rows:       []domainReporting.ARAgingRow{{Bucket: "CURRENT", InvoiceCount: 2, Outstanding: 150000}},
		TotalCount: 2, Total: 150000,
	}, nil
}

func (f *fakeAdvancedRepo) GatewayReconciliation(_ context.Context, tenant string, from, to time.Time) (domainReporting.GatewayReconciliationReport, error) {
	f.tenant, f.from, f.to = tenant, from, to
	return domainReporting.GatewayReconciliationReport{
		From: from, To: to,
		Rows:         []domainReporting.GatewayReconRow{{Gateway: "TRIPAY", TransactionCount: 1, GatewayAmount: 110000, RecordedAmount: 110000}},
		TotalGateway: 110000, TotalRecorded: 110000,
	}, nil
}

func (f *fakeAdvancedRepo) PSBStats(_ context.Context, tenant string, from, to time.Time) (domainReporting.PSBStatsReport, error) {
	f.tenant, f.from, f.to = tenant, from, to
	return domainReporting.PSBStatsReport{
		From: from, To: to, Total: 3, Converted: 2, Closed: 1,
		ByStatus:       []domainReporting.PSBStatusCount{{Status: "ACTIVE", Count: 2}},
		AvgSurveyHours: 4.5, AvgLeadHours: 26,
	}, nil
}

func (f *fakeAdvancedRepo) TaxReport(_ context.Context, _ string, period string) (domainReporting.TaxReport, error) {
	return domainReporting.TaxReport{Period: period, TotalTax: 11000, TotalBilled: 111000}, nil
}

func (f *fakeAdvancedRepo) SalesReport(_ context.Context, _ string, period string) (domainReporting.SalesReport, error) {
	return domainReporting.SalesReport{Period: period, TotalBilled: 100000, TotalPaid: 90000}, nil
}

func (f *fakeAdvancedRepo) CollectionReport(_ context.Context, _ string, period string) (domainReporting.CollectionReport, error) {
	return domainReporting.CollectionReport{Period: period, CollectionRate: 90}, nil
}

func (f *fakeAdvancedRepo) CashflowReport(_ context.Context, _ string, from, to time.Time) (domainReporting.CashflowReport, error) {
	return domainReporting.CashflowReport{From: from, To: to, TotalIn: 100000, TotalOut: 40000, NetCashflow: 60000}, nil
}

func (f *fakeAdvancedRepo) AgingDetail(_ context.Context, _ string, asOf time.Time, _ int) (domainReporting.AgingDetailReport, error) {
	return domainReporting.AgingDetailReport{AsOf: asOf, TotalCustomers: 1, TotalOutstanding: 50000}, nil
}

func (f *fakeAdvancedRepo) ClosingReport(_ context.Context, _ string, period string, closingDate time.Time) (domainReporting.ClosingReport, error) {
	return domainReporting.ClosingReport{Period: period, ClosingDate: closingDate, TotalBilled: 100000}, nil
}

func TestAdvancedReports_Delegation(t *testing.T) {
	repo := &fakeAdvancedRepo{}
	useCase := uc.NewManageReportUseCase(nil, nil).WithAdvancedRepository(repo)
	ctx := context.Background()

	ar, err := useCase.ARAging(ctx, "2026-09-30")
	require.NoError(t, err)
	assert.Equal(t, "tenant-default", repo.tenant)
	assert.Equal(t, time.Date(2026, 9, 30, 0, 0, 0, 0, time.UTC), repo.asOf)
	assert.Equal(t, 2, ar.TotalCount)
	assert.InDelta(t, 150000, ar.Total, 0.01)

	recon, err := useCase.GatewayReconciliation(ctx, "2026-09-01", "2026-09-30")
	require.NoError(t, err)
	assert.Equal(t, time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC), repo.from)
	assert.Equal(t, time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC), repo.to, "to inklusif → +1 hari")
	assert.InDelta(t, 110000, recon.TotalGateway, 0.01)

	psb, err := useCase.PSBStats(ctx, "2026-09-01", "2026-09-15")
	require.NoError(t, err)
	assert.Equal(t, 2, psb.Converted)
	assert.InDelta(t, 4.5, psb.AvgSurveyHours, 0.01)

	tax, err := useCase.TaxReport(ctx, "2026-09")
	require.NoError(t, err)
	assert.Equal(t, "2026-09", tax.Period)

	sales, err := useCase.SalesReport(ctx, "2026-09")
	require.NoError(t, err)
	assert.InDelta(t, 90000, sales.TotalPaid, 0.01)

	collection, err := useCase.CollectionReport(ctx, "2026-09")
	require.NoError(t, err)
	assert.InDelta(t, 90, collection.CollectionRate, 0.01)

	cashflow, err := useCase.CashflowReport(ctx, "2026-09-01", "2026-09-30")
	require.NoError(t, err)
	assert.InDelta(t, 60000, cashflow.NetCashflow, 0.01)

	aging, err := useCase.AgingDetail(ctx, "2026-09-30", 10)
	require.NoError(t, err)
	assert.Equal(t, 1, aging.TotalCustomers)

	closing, err := useCase.ClosingReport(ctx, "2026-09")
	require.NoError(t, err)
	assert.Equal(t, "2026-09", closing.Period)
	assert.Equal(t, time.Date(2026, 9, 30, 0, 0, 0, 0, time.UTC), closing.ClosingDate)
}

func TestExportReportUnknownType(t *testing.T) {
	repo := &fakeAdvancedRepo{}
	useCase := uc.NewManageReportUseCase(nil, nil).WithAdvancedRepository(repo)
	_, _, _, err := useCase.ExportReport(context.Background(), "UNKNOWN", "csv", nil)
	assert.ErrorIs(t, err, domainReporting.ErrInvalidInput)
}

func TestAdvancedReports_InvalidDates(t *testing.T) {
	useCase := uc.NewManageReportUseCase(nil, nil).WithAdvancedRepository(&fakeAdvancedRepo{})
	ctx := context.Background()

	_, err := useCase.ARAging(ctx, "30-09-2026")
	assert.ErrorIs(t, err, domainReporting.ErrInvalidInput)

	_, err = useCase.GatewayReconciliation(ctx, "salah", "")
	assert.ErrorIs(t, err, domainReporting.ErrInvalidInput)

	_, err = useCase.PSBStats(ctx, "", "2026/09/30")
	assert.ErrorIs(t, err, domainReporting.ErrInvalidInput)
}

func TestAdvancedReports_Unavailable(t *testing.T) {
	useCase := uc.NewManageReportUseCase(nil, nil)
	_, err := useCase.ARAging(context.Background(), "")
	assert.ErrorIs(t, err, domainReporting.ErrNotFound)
}
