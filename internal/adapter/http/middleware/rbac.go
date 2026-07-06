package middleware

import "context"

// RBAC enforces Casbin-based role checks on incoming requests.
// TODO: implement per TECH-STACK-DAN-PERSIAPAN.md §3 (Casbin v3.10.0 +
// gorm-adapter v3, single-tenant, model superadmin/owner/admin/staff/teknisi).
func RBAC(ctx context.Context) error {
	return nil
}
