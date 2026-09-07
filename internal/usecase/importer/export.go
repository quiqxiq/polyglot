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
	devices port.DeviceRepository
}

// NewExportUseCase membuat use case ekspor pelanggan.
func NewExportUseCase(
	subs port.SubscriptionRepository,
	customs port.CustomerRepository,
	plans port.ServicePlanRepository,
	devices port.DeviceRepository,
) *ExportUseCase {
	return &ExportUseCase{subs: subs, customs: customs, plans: plans, devices: devices}
}

// ExportAll mengekspor seluruh pelanggan-langganan dalam format 'csv'|'xlsx',
// dengan filter router opsional via deviceIDFilter.
func (u *ExportUseCase) ExportAll(ctx context.Context, format string, deviceIDFilter ...string) ([]byte, error) {
	filter := ""
	if len(deviceIDFilter) > 0 {
		filter = deviceIDFilter[0]
	}
	rows, err := u.CollectRows(ctx, filter)
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
func (u *ExportUseCase) CollectRows(ctx context.Context, deviceIDFilter ...string) ([]Row, error) {
	filter := ""
	if len(deviceIDFilter) > 0 {
		filter = deviceIDFilter[0]
	}
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
	planPrices := map[string]float64{}
	if plans, err := u.plans.List(ctx, false); err == nil {
		for _, pl := range plans {
			planNames[pl.ID] = pl.Name
			planPrices[pl.ID] = pl.Price
		}
	}
	deviceNames := map[string]string{}
	if u.devices != nil {
		if devs, err := u.devices.FindAll(ctx); err == nil {
			for _, d := range devs {
				deviceNames[d.ID] = d.Name
			}
		}
	}

	sort.Slice(subs, func(i, j int) bool { return subs[i].CreatedAt.Before(subs[j].CreatedAt) })
	rows := make([]Row, 0, len(subs))
	for _, s := range subs {
		if filter != "" && (s.DeviceID == nil || *s.DeviceID != filter) {
			continue
		}
		cust := byID[s.CustomerID]
		devName := ""
		if s.DeviceID != nil {
			if name, ok := deviceNames[*s.DeviceID]; ok && name != "" {
				devName = name
			} else {
				devName = *s.DeviceID
			}
		}
		price := planPrices[s.PlanID]
		r := Row{
			CustomerCode: cust.CustomerCode,
			Name:         cust.Name,
			Phone:        cust.Phone,
			Email:        cust.Email,
			Address:      cust.Address,
			Latitude:     cust.Latitude,
			Longitude:    cust.Longitude,
			ServiceType:  s.ServiceType,
			DeviceName:   devName,
			Username:     s.RemoteUsername,
			PlanName:     planNameOr(s.PlanID, planNames),
			Price:        price,
			RateLimit:    s.RateLimit,
			Status:       s.Status,
			LocalAddress: s.LocalAddress,
			RemoteAddr:   s.RemoteAddress,
			ParentQueue:  s.ParentQueue,
			BillingDay:   s.BillingDay,
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
