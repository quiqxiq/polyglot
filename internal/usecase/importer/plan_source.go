package importer

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/quixiq/polyglot/internal/port"
)

// PlanRow merepresentasikan satu baris profil router yang siap diimpor sebagai ServicePlan.
type PlanRow struct {
	Name                  string // Nama komersial paket (editable)
	ServiceType           string // "PPPOE" | "HOTSPOT"
	RateLimit             string // "5M/10M"
	BandwidthDownloadKbps int
	BandwidthUploadKbps   int
	Price                 float64
	ParentQueue           string
	AddressList           string
	IPPoolName            string
	SharedUsers           int
	SessionTimeout        string
	IdleTimeout           string
	IsNew                 bool
	Selected              bool
	RouterProfile         string // Nama profil asli di router MikroTik (read-only)
}

// PullRouterPlans mengambil profil PPP dan/atau Hotspot dari router.
func (s *RouterSource) PullRouterPlans(
	ctx context.Context,
	driver port.DeviceDriver,
	serviceType string,
) ([]PlanRow, int, int, error) {
	st := strings.ToUpper(strings.TrimSpace(serviceType))
	rows := make([]PlanRow, 0)
	pppoeDetected := 0
	hotspotDetected := 0

	// 1. PPPoE Profiles (/ppp/profile)
	if (st == "" || st == "ALL" || st == "PPPOE") && s.gateway != nil {
		profiles, err := s.gateway.ListProfiles(ctx, driver, "")
		if err != nil {
			return nil, 0, 0, fmt.Errorf("list ppp profiles: %w", err)
		}
		for _, p := range profiles {
			name := strings.TrimSpace(p.Name)
			if name == "" || name == "default-encryption" {
				continue
			}
			dl, ul := splitRate(p.RateLimit)
			price := extractPriceFromComment(p.Comment)
			isNew := true
			if s.plans != nil {
				if _, err := s.plans.FindByName(ctx, "tenant-default", name); err == nil {
					isNew = false
				}
			}
			rows = append(rows, PlanRow{
				Name:                  FormatAutoName(name),
				ServiceType:           "PPPOE",
				RateLimit:             p.RateLimit,
				BandwidthDownloadKbps: dl,
				BandwidthUploadKbps:   ul,
				Price:                 price,
				ParentQueue:           p.ParentQueue,
				AddressList:           p.AddressList,
				IPPoolName:            p.RemoteAddress,
				SharedUsers:           1,
				IsNew:                 isNew,
				Selected:              true,
				RouterProfile:         name,
			})
			pppoeDetected++
		}
	}

	// 2. Hotspot Profiles (/ip/hotspot/user/profile)
	if (st == "" || st == "ALL" || strings.HasPrefix(st, "HOTSPOT")) && s.hotspotGw != nil {
		profiles, err := s.hotspotGw.GetUserProfiles(ctx, driver)
		if err != nil {
			return nil, 0, 0, fmt.Errorf("list hotspot user profiles: %w", err)
		}
		for _, p := range profiles {
			name := strings.TrimSpace(p.Name)
			if name == "" {
				continue
			}
			dl, ul := splitRate(p.RateLimit)
			price := extractPriceFromComment(p.Comment)
			sharedUsers := 1
			if su, err := strconv.Atoi(p.SharedUsers); err == nil && su > 0 {
				sharedUsers = su
			}
			isNew := true
			if s.plans != nil {
				if _, err := s.plans.FindByName(ctx, "tenant-default", name); err == nil {
					isNew = false
				}
			}
			rows = append(rows, PlanRow{
				Name:                  FormatAutoName(name),
				ServiceType:           "HOTSPOT",
				RateLimit:             p.RateLimit,
				BandwidthDownloadKbps: dl,
				BandwidthUploadKbps:   ul,
				Price:                 price,
				ParentQueue:           p.ParentQueue,
				AddressList:           p.AddressList,
				IPPoolName:            p.AddressPool,
				SharedUsers:           sharedUsers,
				SessionTimeout:        p.SessionTimeout,
				IdleTimeout:           p.IdleTimeout,
				IsNew:                 isNew,
				Selected:              true,
				RouterProfile:         name,
			})
			hotspotDetected++
		}
	}

	return rows, pppoeDetected, hotspotDetected, nil
}
