package network

import (
	"context"
	"fmt"

	"github.com/quixiq/polyglot/internal/domain/device"
	"github.com/quixiq/polyglot/internal/driver/genericssh"
	"github.com/quixiq/polyglot/internal/port"
)

// OpenTerminalUseCase orchestrates opening a real interactive PTY session (SSH/Telnet) to a target network device.
type OpenTerminalUseCase struct {
	repo  port.DeviceRepository
	vault port.CredentialVault
}

// NewOpenTerminalUseCase constructs a new OpenTerminalUseCase.
func NewOpenTerminalUseCase(repo port.DeviceRepository, vault port.CredentialVault) *OpenTerminalUseCase {
	return &OpenTerminalUseCase{
		repo:  repo,
		vault: vault,
	}
}

// Execute resolves device inventory and credentials, constructs target, and opens an SSH PTY session.
func (u *OpenTerminalUseCase) Execute(ctx context.Context, deviceID string, cols, rows int) (port.TerminalSession, error) {
	if deviceID == "" {
		return nil, fmt.Errorf("open_terminal: device id is required")
	}

	dev, err := u.repo.FindByID(ctx, deviceID)
	if err != nil {
		return nil, fmt.Errorf("open_terminal: find device %s: %w", deviceID, err)
	}

	creds, err := u.vault.Get(ctx, deviceID)
	if err != nil {
		// Fallback to default credentials if not stored in vault
		creds = device.Credentials{
			Username: "admin",
			Password: "r00t",
		}
	}

	target := dev.ToTarget(creds)
	return genericssh.DialSSHPty(ctx, target, cols, rows)
}
