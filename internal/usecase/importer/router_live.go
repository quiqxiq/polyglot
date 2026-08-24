package importer

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/quixiq/polyglot/internal/port"
)

// extractPriceFromComment mencoba membaca harga dari komentar secret
// (konvensi Mikhmon menyimpan metadata pada comment).
func extractPriceFromComment(comment string) float64 {
	for _, part := range strings.Fields(comment) {
		if strings.HasPrefix(strings.ToLower(part), "rp") {
			clean := strings.NewReplacer("Rp", "", "rp", "", ".", "").Replace(part)
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
	gateway port.PPPGateway
}

func NewRouterSource(gw port.PPPGateway) *RouterSource { return &RouterSource{gateway: gw} }

// PullRows mengubah seluruh PPP secret pada device menjadi Row impor.
// Hotspot users ditangani pemanggil bila diperlukan (fase lanjut).
func (s *RouterSource) PullPPPoERows(ctx context.Context, driver port.DeviceDriver, deviceName string) ([]Row, error) {
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
		_ = extractPriceFromComment(comment) // harga dari comment (konvensi Mikhmon)
		rateLimit := extractRateFromProfileHint(comment)
		rows = append(rows, Row{
			Name:        name, // Mikhmon konvensi: nama = username
			Phone:       guessPhone(comment),
			Address:     "",
			ServiceType: "PPPOE",
			DeviceName:  deviceName,
			Username:    name,
			// RouterOS tidak mengekspor password secret via API print;
			// biarkan kosong — admin reset via portal bila perlu.
			PlanName:     orValue(sec.Profile, "UNKNOWN"),
			RateLimit:    rateLimit,
			Status:       status,
			LocalAddress: sec.LocalAddress,
			RemoteAddr:   sec.RemoteAddress,
			RowNumber:    len(rows) + 2,
		})
	}
	return rows, nil
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
