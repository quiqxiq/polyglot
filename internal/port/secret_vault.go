package port

import (
	"context"
)

// SecretVault stores small sensitive strings (account passwords for
// provisioned PPPoE secrets / permanent hotspot users) encrypted at rest.
// It is intentionally generic — unlike CredentialVault which is bound to
// device credentials — and is keyed by a logical key, convention:
// "subscription:<subscriptionID>:password".
type SecretVault interface {
	Put(ctx context.Context, key, secret string) error
	Get(ctx context.Context, key string) (string, error)
	Delete(ctx context.Context, key string) error
}
