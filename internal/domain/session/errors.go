package session

import "github.com/quixiq/polyglot/pkg/fault"

// Sentinel errors for the session domain (tokens, refresh flow).
var (
	// ErrRefreshTokenRequired indicates the request did not carry a refresh token.
	ErrRefreshTokenRequired = fault.New(fault.KindInvalidInput, "session: refresh token is required")
	// ErrInvalidRefreshToken indicates the presented refresh token is unknown,
	// expired, or already rotated.
	ErrInvalidRefreshToken = fault.New(fault.KindUnauthenticated, "session: invalid or expired refresh token")
	// ErrInvalidToken indicates the presented access token is invalid, expired, or malformed.
	ErrInvalidToken = fault.New(fault.KindUnauthenticated, "session: invalid or expired token")
	// ErrRefreshUnavailable indicates the refresh token service is not
	// configured in this deployment.
	ErrRefreshUnavailable = fault.New(fault.KindUnavailable, "session: refresh token service not available")
)
