package port

import (
	"context"

	"github.com/quixiq/polyglot/internal/domain/device"
)

// CredentialVault defines access to encrypted secrets: device credentials
// and sensitive subscription data. AI/LLM layers must never see raw
// credentials — only this interface's implementation
// (internal/adapter/postgres.CredentialVault, AES-GCM) may read them, per
// Polyglot-Architecture.md §2 prinsip 4. The vault decrypts the
// credentials blob (see migration 000001's credentials table) and returns
// a device.Credentials struct, which is merged into a device.Target by
// Device.ToTarget before being passed to a vendor's NewDriver.
type CredentialVault interface {
	// Get returns the decrypted credentials for deviceID, or an error
	// wrapping device.ErrNotFound if no credential row exists.
	Get(ctx context.Context, deviceID string) (device.Credentials, error)

	// Save encrypts and stores credentials for deviceID.
	Save(ctx context.Context, deviceID string, creds device.Credentials) error

	// EncryptString seals an arbitrary sensitive string (AES-GCM, base64)
	// menggunakan key vault yang sama — dipakai mis. untuk
	// subscriptions.remote_password_cipher.
	EncryptString(ctx context.Context, plaintext string) (string, error)

	// DecryptString opens a ciphertext produced by EncryptString.
	DecryptString(ctx context.Context, ciphertext string) (string, error)
}
