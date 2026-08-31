package provisioner

import (
	"context"
	"strings"

	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/internal/registry"
)

// Provisioner implements port.RouterAccountManager by composing
// driver-resolver + PPPGateway + HotspotGateway + FirewallGateway + QueueGateway.
// It enforces realistic ISP subscriber lifecycles on target routers:
//   - isolate = switch profile + kick active sessions + address-list redirect
//   - suspend = disable account + kick
//   - terminate = remove account + queue
type Provisioner struct {
	resolve func(ctx context.Context, deviceID string) (port.DeviceDriver, error)
	ppp     port.PPPGateway
	hot     port.HotspotGateway
	fw      port.FirewallGateway
	q       port.QueueGateway
}

var _ port.RouterAccountManager = (*Provisioner)(nil)

// New creates a new Provisioner backed by the given registry and gateways.
func New(
	reg *registry.Registry,
	ppp port.PPPGateway,
	hot port.HotspotGateway,
	fw port.FirewallGateway,
	q port.QueueGateway,
) *Provisioner {
	return &Provisioner{
		resolve: func(ctx context.Context, deviceID string) (port.DeviceDriver, error) {
			return reg.Get(ctx, deviceID)
		},
		ppp: ppp,
		hot: hot,
		fw:  fw,
		q:   q,
	}
}

// NewWithResolver creates a Provisioner with a custom driver resolver function (useful for tests).
func NewWithResolver(
	resolve func(ctx context.Context, deviceID string) (port.DeviceDriver, error),
	ppp port.PPPGateway,
	hot port.HotspotGateway,
	fw port.FirewallGateway,
	q port.QueueGateway,
) *Provisioner {
	return &Provisioner{
		resolve: resolve,
		ppp:     ppp,
		hot:     hot,
		fw:      fw,
		q:       q,
	}
}

const serviceHotspot = "HOTSPOT"

func isHotspot(serviceType string) bool {
	return strings.EqualFold(serviceType, serviceHotspot)
}
