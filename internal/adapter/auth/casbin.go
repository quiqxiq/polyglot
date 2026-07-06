package auth

import "context"

// NewCasbin builds the Casbin enforcer (single-tenant RBAC, model
// superadmin/owner/admin/staff/teknisi, no domain/tenant scoping — see
// NetOps-Architecture.md §12 keputusan tenancy).
// TODO: use github.com/casbin/casbin/v3 (v3.10.0) with
// github.com/casbin/gorm-adapter/v3 to store policy in the existing GORM/
// Postgres connection, per TECH-STACK-DAN-PERSIAPAN.md §3.
func NewCasbin(ctx context.Context) error {
	return nil
}
