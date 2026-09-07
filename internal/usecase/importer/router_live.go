package importer

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/quixiq/polyglot/internal/port"
)

// extractPriceFromComment mencoba membaca harga dari komentar secret/user/binding
// (konvensi Mikhmon / ISP menyimpan metadata harga pada comment).
func extractPriceFromComment(comment string) float64 {
	normalized := strings.ReplaceAll(comment, "Rp ", "Rp")
	normalized = strings.ReplaceAll(normalized, "rp ", "rp")
	normalized = strings.ReplaceAll(normalized, "RP ", "RP")
	for _, part := range strings.Fields(normalized) {
		lower := strings.ToLower(part)
		if strings.HasPrefix(lower, "rp") {
			clean := strings.NewReplacer("rp.", "", "rp", "", ".", "", ",", ".").Replace(lower)
			if v, err := strconv.ParseFloat(clean, 64); err == nil {
				return v
			}
		}
	}
	return 0
}

// extractRateFromProfileHint tidak tersedia di level secret (rate ada di
// profil); dikembalikan kosong agar mengikuti profil paket.
func extractRateFromProfileHint(string) string { return "" }

// guessPhone mencari pola nomor di komentar.
func guessPhone(comment string) string {
	for _, part := range strings.Fields(comment) {
		digits := strings.Map(func(r rune) rune {
			if r >= '0' && r <= '9' || r == '+' {
				return r
			}
			return -1
		}, part)
		if len(digits) >= 10 && strings.HasPrefix(digits, "08") ||
			len(digits) >= 11 && strings.HasPrefix(digits, "628") {
			return digits
		}
	}
	return ""
}

// RouterSource membaca akun pelanggan langsung dari router (mikrotik gateway).
type RouterSource struct {
	gateway   port.PPPGateway
	hotspotGw port.HotspotGateway
}

// NewRouterSource menginstansiasi router source untuk menarik data PPPoE dan Hotspot.
func NewRouterSource(gw port.PPPGateway, hotspotGw port.HotspotGateway) *RouterSource {
	return &RouterSource{gateway: gw, hotspotGw: hotspotGw}
}

// PullOptions parameter penyaringan penarikan akun dari router.
type PullOptions struct {
	ServiceType       string // "PPPOE", "HOTSPOT", "ALL"
	IncludeIPBindings bool   // tarik /ip/hotspot/ip-binding
	IncludeVouchers   bool   // tarik voucher sementara (default: false)
}

// PullResult hasil pembacaan akun dari router beserta rincian deteksi.
type PullResult struct {
	Rows                     []Row
	PPPoEDetected            int
	HotspotPermanentDetected int
	HotspotIPBindingDetected int
	VouchersSkipped          int
}

// IsVoucherProfile mendeteksi apakah profil hotspot merupakan profil voucher Mikhmon
// berdasarkan skrip on-login atau pola validity pada komentar profil.
func IsVoucherProfile(p port.HotspotUserProfile) bool {
	combined := strings.ToLower(p.OnLogin)
	if strings.Contains(combined, "mikhmon") ||
		strings.Contains(combined, "fetch") ||
		strings.Contains(combined, "scheduler") ||
		strings.Contains(combined, "remov") ||
		strings.Contains(combined, "expire") {
		return true
	}
	comment := strings.ToLower(p.Comment)
	return strings.Contains(comment, "validity") || strings.Contains(comment, "exp")
}

// PullPPPoERows mengubah seluruh PPP secret pada device menjadi Row impor.
func (s *RouterSource) PullPPPoERows(ctx context.Context, driver port.DeviceDriver, deviceName string) ([]Row, error) {
	if s.gateway == nil {
		return nil, nil
	}
	secrets, err := s.gateway.ListSecrets(ctx, driver, "")
	if err != nil {
		return nil, fmt.Errorf("list secrets: %w", err)
	}
	rows := make([]Row, 0, len(secrets))
	for _, sec := range secrets {
		name := strings.TrimSpace(sec.Name)
		if name == "" || strings.HasPrefix(name, "e2e-") {
			continue // buang artefak test
		}
		status := "ACTIVE"
		if sec.Disabled {
			status = "SUSPENDED"
		}
		comment := sec.Comment
		price := extractPriceFromComment(comment)
		rateLimit := extractRateFromProfileHint(comment)
		rows = append(rows, Row{
			Name:         name, // Mikhmon konvensi: nama = username
			Phone:        guessPhone(comment),
			Address:      "",
			ServiceType:  "PPPOE",
			DeviceName:   deviceName,
			Username:     name,
			PlanName:     orValue(sec.Profile, "UNKNOWN"),
			Price:        price,
			RateLimit:    rateLimit,
			Status:       status,
			LocalAddress: sec.LocalAddress,
			RemoteAddr:   sec.RemoteAddress,
			RowNumber:    len(rows) + 2,
		})
	}
	return rows, nil
}

// PullHotspotRows membaca akun hotspot permanen dan IP binding dari router.
func (s *RouterSource) PullHotspotRows(
	ctx context.Context,
	driver port.DeviceDriver,
	deviceName string,
	includeIPBindings bool,
	includeVouchers bool,
) ([]Row, int, int, int, error) {
	if s.hotspotGw == nil {
		return nil, 0, 0, 0, nil
	}

	// 1. Ambil profil untuk memeriksa kehadiran skrip Mikhmon
	profiles, _ := s.hotspotGw.GetUserProfiles(ctx, driver)
	voucherProfiles := make(map[string]bool, len(profiles))
	for _, p := range profiles {
		if IsVoucherProfile(p) {
			voucherProfiles[p.Name] = true
		}
	}

	rows := make([]Row, 0)
	permUsersDetected := 0
	ipBindingsDetected := 0
	vouchersSkipped := 0

	// 2. Tarik IP Binding (Static IP / Bypassed devices) bila diminta
	if includeIPBindings {
		bindings, err := s.hotspotGw.ListIPBindings(ctx, driver)
		if err == nil {
			for _, b := range bindings {
				if b.Disabled {
					continue
				}
				name := strings.TrimSpace(b.Comment)
				if name == "" {
					name = b.MACAddress
				}
				if name == "" {
					name = b.Address
				}
				if name == "" {
					continue
				}
				rows = append(rows, Row{
					Name:         name,
					Phone:        guessPhone(b.Comment),
					Address:      "",
					ServiceType:  "HOTSPOT",
					DeviceName:   deviceName,
					Username:     name,
					Password:     "",
					PlanName:     "IP-BINDING",
					Price:        extractPriceFromComment(b.Comment),
					Status:       "ACTIVE",
					LocalAddress: b.ToAddress,
					RemoteAddr:   b.Address,
					MACAddress:   b.MACAddress,
					HotspotType:  "IP_BINDING",
					RowNumber:    len(rows) + 2,
				})
				ipBindingsDetected++
			}
		}
	}

	// 3. Tarik Hotspot Users
	users, err := s.hotspotGw.ListUsers(ctx, driver, port.ListUsersFilter{})
	if err != nil {
		return rows, permUsersDetected, ipBindingsDetected, vouchersSkipped, fmt.Errorf("list hotspot users: %w", err)
	}

	for _, u := range users {
		name := strings.TrimSpace(u.Name)
		if name == "" || strings.HasPrefix(name, "e2e-") {
			continue
		}

		// Deteksi apakah user merupakan voucher ephemeral
		isVoucher := voucherProfiles[u.Profile] ||
			u.LimitUptime != "" ||
			u.LimitBytesIn != "" ||
			u.LimitBytesOut != "" ||
			strings.HasPrefix(strings.ToLower(u.Comment), "vc-") ||
			strings.Contains(strings.ToLower(u.Comment), "exp:")

		if isVoucher {
			vouchersSkipped++
			if !includeVouchers {
				continue // lewati voucher sementara
			}
		} else {
			permUsersDetected++
		}

		status := "ACTIVE"
		if u.Disabled {
			status = "SUSPENDED"
		}
		comment := u.Comment
		price := extractPriceFromComment(comment)
		hotspotType := "PERMANENT_USER"
		if isVoucher {
			hotspotType = "VOUCHER"
		}

		rows = append(rows, Row{
			Name:        name,
			Phone:       guessPhone(comment),
			Address:     "",
			ServiceType: "HOTSPOT",
			DeviceName:  deviceName,
			Username:    name,
			Password:    u.Password,
			PlanName:    orValue(u.Profile, "default"),
			Price:       price,
			Status:      status,
			RemoteAddr:  u.Address,
			MACAddress:  u.MACAddress,
			HotspotType: hotspotType,
			RowNumber:   len(rows) + 2,
		})
	}

	return rows, permUsersDetected, ipBindingsDetected, vouchersSkipped, nil
}

// PullRouterRows membaca seluruh akun (PPPoE dan/atau Hotspot) sesuai opsi yang dipilih.
func (s *RouterSource) PullRouterRows(
	ctx context.Context,
	driver port.DeviceDriver,
	deviceName string,
	opts PullOptions,
) (*PullResult, error) {
	res := &PullResult{Rows: make([]Row, 0)}
	st := strings.ToUpper(strings.TrimSpace(opts.ServiceType))

	// PPPoE
	if st == "" || st == "ALL" || st == "PPPOE" {
		pRows, err := s.PullPPPoERows(ctx, driver, deviceName)
		if err != nil {
			return nil, err
		}
		res.PPPoEDetected = len(pRows)
		res.Rows = append(res.Rows, pRows...)
	}

	// Hotspot
	if st == "" || st == "ALL" || strings.HasPrefix(st, "HOTSPOT") {
		hRows, permDetected, ipbDetected, vSkipped, err := s.PullHotspotRows(
			ctx, driver, deviceName, opts.IncludeIPBindings, opts.IncludeVouchers,
		)
		if err != nil {
			return nil, err
		}
		res.HotspotPermanentDetected = permDetected
		res.HotspotIPBindingDetected = ipbDetected
		res.VouchersSkipped = vSkipped
		res.Rows = append(res.Rows, hRows...)
	}

	// Re-number rows for friendly error message indexing
	for i := range res.Rows {
		res.Rows[i].RowNumber = i + 2
	}

	return res, nil
}

// Reconciler membandingkan DB (langganan provisioned per device) vs router.
type DriftReport struct {
	MissingInDB     []string `json:"missing_in_db"`     // ada di router, tak ada di DB
	MissingInRoute  []string `json:"missing_in_router"` // ada di DB, tak ada di router
	ProfileMismatch []string `json:"profile_mismatch"`
}

func NewReconciler(subs port.SubscriptionRepository, gw port.PPPGateway) *Reconciler {
	return &Reconciler{subs: subs, gw: gw}
}

type Reconciler struct {
	subs port.SubscriptionRepository
	gw   port.PPPGateway
}

// Compare membandingkan username langganan provisioned-OK di DB dengan
// secrets di router untuk deviceID tertentu.
func (r *Reconciler) Compare(ctx context.Context, deviceID string, driver port.DeviceDriver) (*DriftReport, error) {
	report := &DriftReport{}

	lc, err := r.subs.ListLifecycle(ctx)
	if err != nil {
		return nil, err
	}
	dbNames := map[string]string{} // username → router_profile
	for _, s := range lc {
		if s.DeviceID == nil || *s.DeviceID != deviceID || s.RemoteUsername == "" {
			continue
		}
		dbNames[s.RemoteUsername] = s.RouterProfile
	}

	routerProfiles := map[string]string{}
	secrets, err := r.gw.ListSecrets(ctx, driver, "")
	if err != nil {
		return nil, err
	}
	for _, sec := range secrets {
		routerProfiles[sec.Name] = sec.Profile
	}

	for name, profile := range dbNames {
		rp, onRouter := routerProfiles[name]
		if !onRouter {
			report.MissingInRoute = append(report.MissingInRoute, name)
			continue
		}
		if profile != "" && rp != "" && rp != profile {
			report.ProfileMismatch = append(report.ProfileMismatch,
				fmt.Sprintf("%s: db=%s router=%s", name, profile, rp))
		}
	}
	for name := range routerProfiles {
		if _, inDB := dbNames[name]; !inDB && !strings.HasPrefix(name, "e2e-") {
			report.MissingInDB = append(report.MissingInDB, name)
		}
	}
	return report, nil
}
