package port

import "context"

// CredentialVault defines access to encrypted device credentials.
// AI/LLM layers must never see raw credentials — only this interface's
// implementations (internal/adapter/vault) may read them, per
// NetOps-Architecture.md §2 prinsip 4.
type CredentialVault interface {
	Get(ctx context.Context, deviceID string) error
}
