package report

import (
	"encoding/json"

	devicepb "github.com/quixiq/polyglot/api/gen/v1"
	domainReporting "github.com/quixiq/polyglot/internal/domain/reporting"
)

type doubleAlias = float64

func toProtoSummary(period string, snaps []domainReporting.DailyFinancialSnapshot) *devicepb.ReportSummary {
	sum := &devicepb.ReportSummary{
		Period: period, CashBalances: map[string]doubleAlias{},
	}
	for _, s := range snaps {
		sum.InvoiceCount += int32(s.InvoiceCount)
		sum.InvoiceTotal += s.InvoiceTotal
		sum.PaymentCount += int32(s.PaymentCount)
		sum.PaymentTotal += s.PaymentTotal
		sum.OutstandingTotal = s.OutstandingTotal
		sum.ExpenseTotal += s.ExpenseTotal
		sum.ActiveSubscriptions = int32(s.ActiveSubscriptions)
		if len(s.CashBalanceJSON) > 0 {
			var m map[string]float64
			if err := json.Unmarshal(s.CashBalanceJSON, &m); err == nil {
				sum.CashBalances = m
			}
		}
	}
	return sum
}
