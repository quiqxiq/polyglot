package port

import "context"

// Identity is the authenticated principal carried by a validated token. It
// is intentionally free of any token-library type (no JWT claims embedded)
// so this contract layer stays independent of how tokens are implemented.
type Identity struct {
	UserID   string
	Username string
	Role     string
}

// Authenticator issues signed tokens and validates them back into an
// Identity. The concrete implementation (JWT/HMAC, etc.) lives in an
// adapter; callers depend only on this contract.
type Authenticator interface {
	// Issue returns a signed token encoding the given identity fields.
	Issue(ctx context.Context, userID, username, role string) (string, error)
	// Validate verifies token and returns the Identity it encodes, or an
	// error if the token is invalid, expired, or malformed.
	Validate(ctx context.Context, token string) (*Identity, error)
}
