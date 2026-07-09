package port

import (
	"context"

	"github.com/quixiq/polyglot/internal/domain/device"
)

// CredentialVault defines access to encrypted device credentials.
// AI/LLM layers must never see raw credentials — only this interface's
// implementations (internal/adapter/vault) may read them, per
// Polyglot-Architecture.md §2 prinsip 4. The vault decrypts the
// credentials blob (see migration 000001's credentials table) and returns
// a device.Credentials struct, which is merged into a device.Target by
// Device.ToTarget before being passed to a vendor's NewDriver.
type CredentialVault interface {
	// Get returns the decrypted credentials for deviceID, or an error
	// wrapping device.ErrNotFound if no credential row exists.
	Get(ctx context.Context, deviceID string) (device.Credentials, error)
}
