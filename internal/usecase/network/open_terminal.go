package network

import (
	"context"
	"fmt"

	"github.com/quixiq/polyglot/internal/domain/device"
	"github.com/quixiq/polyglot/internal/port"
)

// TerminalDialer dials a target device and establishes an interactive TerminalSession.
type TerminalDialer func(ctx context.Context, target device.Target, cols, rows int) (port.TerminalSession, error)

// OpenTerminalUseCase orchestrates opening a real interactive PTY session (SSH/Telnet) to a target network device.
type OpenTerminalUseCase struct {
	repo   port.DeviceRepository
	vault  port.CredentialVault
	dialer TerminalDialer
}

// NewOpenTerminalUseCase constructs a new OpenTerminalUseCase.
func NewOpenTerminalUseCase(repo port.DeviceRepository, vault port.CredentialVault, dialer TerminalDialer) *OpenTerminalUseCase {
	return &OpenTerminalUseCase{
		repo:   repo,
		vault:  vault,
		dialer: dialer,
	}
}

// Execute resolves device inventory and credentials, constructs target, and opens an SSH PTY session.
func (u *OpenTerminalUseCase) Execute(ctx context.Context, deviceID string, cols, rows int) (port.TerminalSession, error) {
	if deviceID == "" {
		return nil, fmt.Errorf("%w: device id is required", device.ErrInvalidInput)
	}
	if u.dialer == nil {
		return nil, fmt.Errorf("%w: terminal dialer not configured", device.ErrDiagnosticsUnconfigured)
	}

	dev, err := u.repo.FindByID(ctx, deviceID)
	if err != nil {
		return nil, fmt.Errorf("open_terminal: find device %s: %w", deviceID, err)
	}

	creds, err := u.vault.Get(ctx, deviceID)
	if err != nil {
		return nil, fmt.Errorf("open_terminal: get credentials for device %s: %w", deviceID, err)
	}

	target := dev.ToTarget(creds)
	return u.dialer(ctx, target, cols, rows)
}
