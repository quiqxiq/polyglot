// Package oltmanager menjembatani driver vendor OLT ke port.OLTManager:
// usecase hanya melihat operasi terstruktur, bukan perintah CLI/driver.
package oltmanager

import (
	"context"
	"fmt"

	"github.com/quixiq/polyglot/internal/domain/olt"
	"github.com/quixiq/polyglot/internal/driver/hioso"
	"github.com/quixiq/polyglot/internal/port"
)

// DriverResolver mengambil driver ter-cache untuk satu device (registry.Get).
type DriverResolver func(ctx context.Context, deviceID string) (port.DeviceDriver, error)

// Provider membuat OLTManager terikat ke device OLT HiOSO.
type Provider struct {
	resolve DriverResolver
}

// NewProvider membangun provider OLT manager HiOSO.
func NewProvider(resolve DriverResolver) *Provider {
	return &Provider{resolve: resolve}
}

var _ port.OLTManagerProvider = (*Provider)(nil)

// ForDevice me-resolve driver hioso untuk oltDeviceID.
func (p *Provider) ForDevice(ctx context.Context, oltDeviceID string) (port.OLTManager, error) {
	drv, err := p.resolve(ctx, oltDeviceID)
	if err != nil {
		return nil, fmt.Errorf("oltmanager: resolve device %s: %w", oltDeviceID, err)
	}
	h, ok := drv.(*hioso.Driver)
	if !ok {
		return nil, fmt.Errorf("%w: device %s bukan driver hioso", olt.ErrUnsupportedVariant, oltDeviceID)
	}
	return &hiosoManager{drv: h}, nil
}

// hiosoManager mengimplementasikan port.OLTManager di atas *hioso.Driver.
type hiosoManager struct {
	drv *hioso.Driver
}

var _ port.OLTManager = (*hiosoManager)(nil)

func (m *hiosoManager) Detect(ctx context.Context) (olt.Capabilities, port.OLTHealth, error) {
	return m.drv.Detect(ctx)
}

func (m *hiosoManager) ListONUs(ctx context.Context, ponIndex int) ([]port.OLTONU, error) {
	return m.drv.ListONUs(ctx, ponIndex)
}

func (m *hiosoManager) RenameONU(ctx context.Context, ponIndex, onuIndex int, name string) error {
	return m.drv.RenameONU(ctx, ponIndex, onuIndex, name)
}

func (m *hiosoManager) RebootONU(ctx context.Context, ponIndex, onuIndex int) error {
	return m.drv.RebootONU(ctx, ponIndex, onuIndex)
}

func (m *hiosoManager) SetONUAdmin(ctx context.Context, ponIndex, onuIndex int, enabled bool) error {
	return m.drv.SetONUAdmin(ctx, ponIndex, onuIndex, enabled)
}

func (m *hiosoManager) DeleteONU(ctx context.Context, ponIndex, onuIndex int) error {
	return m.drv.DeleteONU(ctx, ponIndex, onuIndex)
}

func (m *hiosoManager) SaveConfig(ctx context.Context) error {
	return m.drv.SaveConfig(ctx)
}

func (m *hiosoManager) RunningConfig(ctx context.Context) (string, error) {
	return m.drv.RunningConfig(ctx)
}

func (m *hiosoManager) SystemInfo(ctx context.Context) (map[string]string, error) {
	return m.drv.SystemInfo(ctx)
}
