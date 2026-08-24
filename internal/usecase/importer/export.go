package importer

import (
	"context"
	"fmt"
	"sort"

	domainCustomer "github.com/quixiq/polyglot/internal/domain/customer"
	"github.com/quixiq/polyglot/internal/port"
)

// ExportUseCase menarik seluruh langganan + pelanggan + plan dari DB dan
// menyusunnya menjadi baris file ekspor (format sama dengan parser impor).
type ExportUseCase struct {
	subs    port.SubscriptionRepository
	customs port.CustomerRepository
	plans   port.ServicePlanRepository
}

func NewExportUseCase(subs port.SubscriptionRepository, customs port.CustomerRepository, plans port.ServicePlanRepository) *ExportUseCase {
	return &ExportUseCase{subs: subs, customs: customs, plans: plans}
}

// ExportAll mengekspor seluruh pelanggan-langganan dalam format 'csv'|'xlsx'.
func (u *ExportUseCase) ExportAll(ctx context.Context, format string) ([]byte, error) {
	rows, err := u.CollectRows(ctx)
	if err != nil {
		return nil, err
	}
	switch format {
	case "xlsx":
		return WriteXLSX(rows)
	default:
		return WriteCSV(rows)
	}
}

// CollectRows menyusun Row gabungan untuk keperluan export/dry-run.
func (u *ExportUseCase) CollectRows(ctx context.Context) ([]Row, error) {
	subs, err := u.subs.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	customers, err := u.customs.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	byID := map[string]domainCustomer.Customer{}
	for _, c := range customers {
		byID[c.ID] = c
	}
	planNames := map[string]string{}
	if plans, err := u.plans.List(ctx, false); err == nil {
		for _, pl := range plans {
			planNames[pl.ID] = pl.Name
		}
	}

	sort.Slice(subs, func(i, j int) bool { return subs[i].CreatedAt.Before(subs[j].CreatedAt) })
	rows := make([]Row, 0, len(subs))
	for _, s := range subs {
		cust := byID[s.CustomerID]
		r := Row{
			CustomerCode: cust.CustomerCode,
			Name:         cust.Name,
			Phone:        cust.Phone,
			Email:        cust.Email,
			Address:      cust.Address,
			Latitude:     cust.Latitude,
			Longitude:    cust.Longitude,
			ServiceType:  s.ServiceType,
			Username:     s.RemoteUsername,
			PlanName:     planNameOr(s.PlanID, planNames),
			RateLimit:    s.RateLimit,
			Status:       s.Status,
			LocalAddress: s.LocalAddress,
			RemoteAddr:   s.RemoteAddress,
			ParentQueue:  s.ParentQueue,
			RowNumber:    len(rows) + 2,
		}
		rows = append(rows, r)
	}
	return rows, nil
}

func planNameOr(planID string, names map[string]string) string {
	if n, ok := names[planID]; ok {
		return n
	}
	return fmt.Sprintf("PLAN-%s", planID)
}
