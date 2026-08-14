package auth

import (
	"context"
	"fmt"
)

// Context keys for propagating the authenticated identity from the gin RBAC
// middleware into downstream ConnectRPC handlers (via req.HTTP().Context()).
// Handlers that need "who is calling" (audit, ownership checks) read it here.
type ctxKey string

const (
	ctxKeyUserID ctxKey = "auth.user_id"
	ctxKeyRoles  ctxKey = "auth.roles"
)

// WithIdentity stores the authenticated user ID and roles in ctx.
func WithIdentity(ctx context.Context, userID uint, roles []string) context.Context {
	ctx = context.WithValue(ctx, ctxKeyUserID, userID)
	ctx = context.WithValue(ctx, ctxKeyRoles, roles)
	return ctx
}

// IdentityFromContext returns the authenticated user ID and roles, if present.
func IdentityFromContext(ctx context.Context) (uint, []string, bool) {
	userID, ok := ctx.Value(ctxKeyUserID).(uint)
	if !ok {
		return 0, nil, false
	}
	roles, _ := ctx.Value(ctxKeyRoles).([]string)
	return userID, roles, true
}

// IdentityStringFromContext is a convenience for audit logs.
func IdentityStringFromContext(ctx context.Context) string {
	userID, _, ok := IdentityFromContext(ctx)
	if !ok {
		return ""
	}
	return fmt.Sprintf("%d", userID)
}
