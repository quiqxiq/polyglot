package importer

import (
	"context"
	"fmt"
	"strings"
	"time"

	domainCustomer "github.com/quixiq/polyglot/internal/domain/customer"
	domainPlan "github.com/quixiq/polyglot/internal/domain/plan"
	domainSubscription "github.com/quixiq/polyglot/internal/domain/subscription"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/pkg/idgen"
)

// UpsertUseCase menerapkan baris impor ke DB: plan dibuat bila nama baru,
// pelanggan di-upsert per nomor HP, langganan di-upsert per
// (device, username). provision_status=OK karena data bersumber dari
// router/Mikhmon (akun memang sudah hidup di router).
type UpsertUseCase struct {
	plans          port.ServicePlanRepository
	customs        port.CustomerRepository
	subs           port.SubscriptionRepository
	audit          port.AuditLogWriter
	defaultDevice  string // deviceID default bila DeviceName tak cocok
	deviceResolver func(deviceName string) (deviceID string, ok bool)
	now            func() time.Time
}

// SetDeviceResolver memetakan nama server (kolom "server") ke devices.id.
func (u *UpsertUseCase) SetDeviceResolver(fn func(string) (string, bool)) {
	u.deviceResolver = fn
}

func NewUpsertUseCase(
	plans port.ServicePlanRepository,
	customs port.CustomerRepository,
	subs port.SubscriptionRepository,
	auditW port.AuditLogWriter,
	defaultDevice string,
) *UpsertUseCase {
	return &UpsertUseCase{
		plans: plans, customs: customs, subs: subs,
		audit: auditW, defaultDevice: defaultDevice, now: time.Now,
	}
}

// Result rekap satu proses impor.
type Result struct {
	RowsTotal        int      `json:"rows_total"`
	CustomersCreated int      `json:"customers_created"`
	CustomersUpdated int      `json:"customers_updated"`
	SubsCreated      int      `json:"subscriptions_created"`
	PlansCreated     int      `json:"plans_created"`
	Skipped          []string `json:"skipped,omitempty"`
}

func (u *UpsertUseCase) Import(ctx context.Context, rows []Row) (*Result, error) {
	res := &Result{RowsTotal: len(rows)}
	for _, r := range rows {
		if err := u.importRow(ctx, r, res); err != nil {
			res.Skipped = append(res.Skipped, fmt.Sprintf("baris %d (%s): %v", r.RowNumber, orDash(r.Name), err))
		}
	}
	return res, nil
}

func (u *UpsertUseCase) importRow(ctx context.Context, r Row, res *Result) error {
	now := u.now()

	planID, err := u.ensurePlan(ctx, r)
	if err != nil {
		return err
	}

	deviceID := u.defaultDevice
	if r.DeviceName != "" && u.deviceResolver != nil {
		if id, ok := u.deviceResolver(r.DeviceName); ok && id != "" {
			deviceID = id
		}
	}
	// ── Pelanggan: upsert per nomor HP ────────────────────────────────
	cust, _, err := u.upsertCustomer(ctx, r, now, res)
	if err != nil {
		return err
	}
	if strings.TrimSpace(r.Username) == "" {
		return nil // tanpa akun jaringan — cukup data master pelanggan
	}

	// ── Langganan: upsert per (device, username) ──────────────────────
	existing, _ := u.subs.FindByDeviceAndUsername(ctx, deviceID, r.Username)
	status := mapStatus(r.Status)
	price := r.Price

	if existing.ID == "" {
		end := now.AddDate(1, 0, 0)
		sub := domainSubscription.Subscription{
			ID: idgen.New("sub"), TenantID: cust.TenantID,
			CustomerID: cust.ID, PlanID: planID, DeviceID: &deviceID,
			ServiceType:    serviceTypeOf(r.ServiceType),
			RemoteUsername: r.Username, RemotePassword: orValue(r.Password, "ganti123"),
			LocalAddress: r.LocalAddress, RemoteAddress: r.RemoteAddr,
			ParentQueue:     orValue(r.ParentQueue, "none"),
			RateLimit:       r.RateLimit,
			BillingCycle:    domainSubscription.CycleMonthly,
			BillingDay:      clampDay(now.Day()),
			Status:          status,
			StartDate:       dayStart(now),
			EndDate:         &end,
			RouterProfile:   r.PlanName,
			ProvisionStatus: domainSubscription.ProvisionOK, // sudah hidup di router
			CreatedAt:       now, UpdatedAt: now,
		}
		if err := u.subs.Save(ctx, sub); err != nil {
			return err
		}
		res.SubsCreated++
		writeAuditImport(ctx, u.audit, "IMPORT_SUBSCRIPTION", sub.ID)
	} else {
		existing.PlanID = planID
		existing.Status = status
		existing.RateLimit = orValue(r.RateLimit, existing.RateLimit)
		existing.RouterProfile = orValue(r.PlanName, existing.RouterProfile)
		existing.ProvisionStatus = domainSubscription.ProvisionOK
		existing.UpdatedAt = now
		_ = price
		if err := u.subs.Save(ctx, existing); err != nil {
			return err
		}
	}
	return nil
}

func (u *UpsertUseCase) ensurePlan(ctx context.Context, r Row) (string, error) {
	planName := strings.TrimSpace(r.PlanName)
	if planName == "" {
		return "", fmt.Errorf("paket kosong")
	}
	if pl, err := u.plans.FindByName(ctx, "tenant-default", planName); err == nil {
		return pl.ID, nil
	}
	dl, ul := splitRate(r.RateLimit)
	pl := domainPlan.ServicePlan{
		ID: idgen.New("plan"), TenantID: "tenant-default", Name: planName,
		ServiceType:           serviceTypeOf(r.ServiceType),
		BandwidthDownloadKbps: dl, BandwidthUploadKbps: ul,
		Price: r.Price, IsActive: true,
		Validity: "30d", ValidityMode: domainPlan.ValidityCalendar,
		ExpireMode: domainPlan.ExpireNotFiltered, ParentQueue: orValue(r.ParentQueue, "none"),
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	if err := u.plans.Save(ctx, pl); err != nil {
		return "", err
	}
	writeAuditImport(ctx, u.audit, "IMPORT_PLAN", pl.ID)
	return pl.ID, nil
}

func (u *UpsertUseCase) upsertCustomer(ctx context.Context, r Row, now time.Time, res *Result) (domainCustomer.Customer, bool, error) {
	phoneN := normalizePhone(r.Phone)
	if phoneN != "" {
		if c, err := u.customs.FindByPhone(ctx, phoneN); err == nil {
			c.Name = orValue(r.Name, c.Name)
			c.Address = orValue(r.Address, c.Address)
			c.Email = orValue(r.Email, c.Email)
			c.Latitude, c.Longitude = coalesceCoord(c.Latitude, c.Longitude, r.Latitude, r.Longitude)
			if c.CustomerCode == "" && r.CustomerCode != "" {
				c.CustomerCode = r.CustomerCode
			}
			c.UpdatedAt = now
			if err := u.customs.Save(ctx, c); err != nil {
				return domainCustomer.Customer{}, false, err
			}
			res.CustomersUpdated++
			return c, false, nil
		}
	}
	cust := domainCustomer.Customer{
		ID: idgen.New("cust"), TenantID: "tenant-default",
		CustomerCode: orValue(r.CustomerCode, "IMP-"+idgen.Digits(6)),
		Name:         r.Name, Phone: phoneN, Email: r.Email,
		Address: r.Address, Latitude: r.Latitude, Longitude: r.Longitude,
		Status:       domainCustomer.StatusActive,
		RegisteredAt: dayStart(now), CreatedAt: now, UpdatedAt: now,
	}
	if err := u.customs.Save(ctx, cust); err != nil {
		return domainCustomer.Customer{}, false, err
	}
	res.CustomersCreated++
	writeAuditImport(ctx, u.audit, "IMPORT_CUSTOMER", cust.ID)
	return cust, true, nil
}

// resolveDeviceID memetakan nama server → deviceID via callback yang
// diinjeksi (SetDeviceResolver); fallback ke defaultDevice.

// ─── helpers ────────────────────────────────────────────────────────────

func serviceTypeOf(t string) string {
	if strings.EqualFold(t, "HOTSPOT") {
		return "HOTSPOT"
	}
	return "PPPOE"
}

func mapStatus(s string) string {
	switch s {
	case "PENDING", "ISOLATED", "SUSPENDED", "TERMINATED":
		return s
	default:
		return domainSubscription.StatusActive
	}
}

func splitRate(rl string) (dl, ul int) {
	parts := strings.SplitN(strings.ToLower(rl), "/", 2)
	parse := func(tok string) int {
		tok = strings.TrimSpace(tok)
		if tok == "" {
			return 0
		}
		mult := 1000 // kbps default token tanpa satuan dianggap M oleh RouterOS
		num := tok
		if strings.HasSuffix(tok, "m") {
			num = strings.TrimSuffix(tok, "m")
		} else if strings.HasSuffix(tok, "k") {
			mult = 1
			num = strings.TrimSuffix(tok, "k")
		}
		var v int
		if _, err := fmt.Sscanf(num, "%d", &v); err != nil {
			return 0
		}
		return v * mult
	}
	dl = parse(parts[0])
	if len(parts) == 2 {
		ul = parse(parts[1])
	} else {
		ul = dl
	}
	return dl, ul
}

func normalizePhone(p string) string { return strings.TrimPrefix(strings.TrimSpace(p), "+") }
func orDash(s string) string         { return orValue(s, "-") }
func orValue(v, def string) string {
	if strings.TrimSpace(v) == "" {
		return def
	}
	return strings.TrimSpace(v)
}
func clampDay(d int) int {
	if d > 28 {
		return 28
	}
	return d
}
func dayStart(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}
func coalesceCoord(aLat, aLon *float64, bLat, bLon *float64) (*float64, *float64) {
	if aLat == nil {
		aLat = bLat
	}
	if aLon == nil {
		aLon = bLon
	}
	return aLat, aLon
}
func writeAuditImport(ctx context.Context, w port.AuditLogWriter, action, entityID string) {
	if w == nil {
		return
	}
	_ = w.Write(ctx, domainAuditEntry(action, entityID))
}
