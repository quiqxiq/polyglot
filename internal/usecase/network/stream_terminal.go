package network

import (
	"context"
	"fmt"

	"github.com/quixiq/polyglot/internal/port"
)

// StreamTerminalUseCase orchestrates opening interactive PTY SSH/Telnet terminal sessions.
type StreamTerminalUseCase struct {
	repo  port.DeviceRepository
	vault port.CredentialVault
}

// NewStreamTerminalUseCase constructs a new StreamTerminalUseCase.
func NewStreamTerminalUseCase(repo port.DeviceRepository, vault port.CredentialVault) *StreamTerminalUseCase {
	return &StreamTerminalUseCase{
		repo:  repo,
		vault: vault,
	}
}

// OpenTerminal opens a PTY TerminalSession if the driver satisfies port.TerminalDeviceDriver.
func (uc *StreamTerminalUseCase) OpenTerminal(ctx context.Context, driver port.DeviceDriver, deviceID string, cols, rows int) (port.TerminalSession, error) {
	if td, ok := driver.(port.TerminalDeviceDriver); ok {
		return td.OpenTerminalSession(ctx, cols, rows)
	}
	return nil, fmt.Errorf("device driver for %s does not support interactive PTY terminal streaming", deviceID)
}
