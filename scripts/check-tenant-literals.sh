#!/usr/bin/env bash
set -euo pipefail

# Polyglot is a single-tenant deployment. The literal is centralized in
# internal/domain/tenant so it cannot drift back into application code.
if rg -n '"tenant-default"' \
  --glob '*.go' \
  --glob '!**/*_test.go' \
  --glob '!internal/domain/tenant/tenant.go' \
  internal cmd pkg test; then
  printf '%s\n' 'tenant literal found outside internal/domain/tenant/tenant.go; use tenant.DefaultID'
  exit 1
fi

printf '%s\n' 'tenant literal check passed'
