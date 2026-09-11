package billing

import (
	"context"
	"fmt"
	"strings"

	domainBilling "github.com/quixiq/polyglot/internal/domain/billing"
	"github.com/quixiq/polyglot/internal/port"
)

// GatewayRegistry menyimpan seluruh payment gateway yang terdaftar dan
// memilih gateway default dari setting `gw.active` (fallback urutan
// registrasi, hanya yang Enabled) — F4-1.
type GatewayRegistry struct {
	settings port.SettingReader
	byName   map[string]port.PaymentGateway
	order    []string
}

// NewGatewayRegistry mendaftarkan gateway dengan urutan prioritas default.
func NewGatewayRegistry(settings port.SettingReader, gateways ...port.PaymentGateway) *GatewayRegistry {
	r := &GatewayRegistry{
		settings: settings,
		byName:   make(map[string]port.PaymentGateway, len(gateways)),
	}
	for _, g := range gateways {
		if g == nil {
			continue
		}
		name := strings.ToUpper(strings.TrimSpace(g.Name()))
		if _, exists := r.byName[name]; exists {
			continue
		}
		r.byName[name] = g
		r.order = append(r.order, name)
	}
	return r
}

// Names mengembalikan nama gateway terdaftar sesuai urutan prioritas.
func (r *GatewayRegistry) Names() []string {
	out := make([]string, len(r.order))
	copy(out, r.order)
	return out
}

// Get mengembalikan gateway berdasarkan nama (case-insensitive); nama kosong
// berarti gateway default.
func (r *GatewayRegistry) Get(ctx context.Context, name string) (port.PaymentGateway, error) {
	name = strings.ToUpper(strings.TrimSpace(name))
	if name == "" {
		return r.Default(ctx)
	}
	g, ok := r.byName[name]
	if !ok {
		return nil, fmt.Errorf("%w: %s", domainBilling.ErrGatewayNotFound, name)
	}
	return g, nil
}

// Default memilih `gw.active` bila aktif, selain itu gateway pertama yang
// Enabled; error ErrGatewayDisabled bila tidak ada yang aktif.
func (r *GatewayRegistry) Default(ctx context.Context) (port.PaymentGateway, error) {
	if r.settings != nil {
		preferred := strings.ToUpper(strings.TrimSpace(r.settings.GetValue(ctx, "gw.active", "")))
		if g, ok := r.byName[preferred]; ok && g.Enabled(ctx) {
			return g, nil
		}
	}
	for _, name := range r.order {
		if g := r.byName[name]; g.Enabled(ctx) {
			return g, nil
		}
	}
	return nil, fmt.Errorf("%w: tidak ada gateway aktif", domainBilling.ErrGatewayDisabled)
}
