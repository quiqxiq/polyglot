package http

import "context"

// NewRouter builds and returns the REST API router.
// This is the ONLY path AdminPages/OtherUI use for CRUD (device, customer,
// subscription, plan) per Polyglot-Architecture.md §6.2 — never proxied
// through the LibreChat Node backend (see §7.3 Batas Otentikasi).
// TODO: wire Gin v1.12.0 engine + routes per TECH-STACK-DAN-PERSIAPAN.md §2.
func NewRouter(ctx context.Context) error {
	return nil
}
