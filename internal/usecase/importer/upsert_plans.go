package importer

import (
	"context"
	"fmt"
	"strings"

	domainPlan "github.com/quixiq/polyglot/internal/domain/plan"
	"github.com/quixiq/polyglot/pkg/idgen"
)

// CommitPlansResult merepresentasikan rekapitulasi hasil penyimpanan paket layanan.
type CommitPlansResult struct {
	PlansCreated int      `json:"plans_created"`
	PlansUpdated int      `json:"plans_updated"`
	Errors       []string `json:"errors,omitempty"`
}

// CommitPlans mengupsert daftar PlanRow ke repository service_plans.
func (u *UpsertUseCase) CommitPlans(ctx context.Context, rows []PlanRow) (*CommitPlansResult, error) {
	res := &CommitPlansResult{}
	now := u.now()

	for _, r := range rows {
		if !r.Selected {
			continue
		}
		name := strings.TrimSpace(r.Name)
		if name == "" {
			res.Errors = append(res.Errors, "nama paket tidak boleh kosong")
			continue
		}

		st := strings.ToUpper(strings.TrimSpace(r.ServiceType))
		if st == "" {
			st = "PPPOE"
		}

		dl := r.BandwidthDownloadKbps
		ul := r.BandwidthUploadKbps
		if dl == 0 && ul == 0 && r.RateLimit != "" {
			dl, ul = splitRate(r.RateLimit)
		}

		sharedUsers := r.SharedUsers
		if sharedUsers <= 0 {
			sharedUsers = 1
		}

		existing, err := u.plans.FindByName(ctx, "tenant-default", name)
		if err == nil && existing.ID != "" {
			existing.ServiceType = st
			existing.BandwidthDownloadKbps = dl
			existing.BandwidthUploadKbps = ul
			if r.Price > 0 {
				existing.Price = r.Price
			}
			if r.ParentQueue != "" {
				existing.ParentQueue = r.ParentQueue
			}
			if r.AddressList != "" {
				existing.AddressList = r.AddressList
			}
			if r.IPPoolName != "" {
				existing.IPPoolName = r.IPPoolName
				existing.RemoteAddressPool = r.IPPoolName
			}
			existing.SharedUsers = sharedUsers
			if r.SessionTimeout != "" {
				existing.SessionTimeout = r.SessionTimeout
			}
			if r.IdleTimeout != "" {
				existing.IdleTimeout = r.IdleTimeout
			}
			existing.UpdatedAt = now
			if err := u.plans.Save(ctx, existing); err != nil {
				res.Errors = append(res.Errors, fmt.Sprintf("paket %q: %v", name, err))
				continue
			}
			res.PlansUpdated++
			writeAuditImport(ctx, u.audit, "IMPORT_PLAN_UPDATE", existing.ID)
		} else {
			newPlan := domainPlan.ServicePlan{
				ID:                    idgen.New("plan"),
				TenantID:              "tenant-default",
				Name:                  name,
				ServiceType:           st,
				BandwidthDownloadKbps: dl,
				BandwidthUploadKbps:   ul,
				Price:                 r.Price,
				ParentQueue:           orValue(r.ParentQueue, "none"),
				AddressList:           r.AddressList,
				IPPoolName:            r.IPPoolName,
				RemoteAddressPool:     r.IPPoolName,
				SharedUsers:           sharedUsers,
				SessionTimeout:        r.SessionTimeout,
				IdleTimeout:           r.IdleTimeout,
				IsActive:              true,
				CreatedAt:             now,
				UpdatedAt:             now,
			}
			if err := u.plans.Save(ctx, newPlan); err != nil {
				res.Errors = append(res.Errors, fmt.Sprintf("paket %q: %v", name, err))
				continue
			}
			res.PlansCreated++
			writeAuditImport(ctx, u.audit, "IMPORT_PLAN_CREATE", newPlan.ID)
		}
	}

	return res, nil
}
