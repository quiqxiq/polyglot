#!/usr/bin/env bash
set -euo pipefail

# Ratchet for legacy type aliases in internal/port.
#
# internal/port historically re-exported domain types (for example
# `type DHCPLease = device.DHCPLease`). Those aliases blur the layer boundary:
# consumers should import the owning domain package directly. Removing all of
# them at once is too risky, so this script freezes the current set:
#
#   * a new alias fails the check immediately;
#   * an alias removed from code must also be removed from the allowlist,
#     which guarantees the allowlist can only shrink.
#
# When removing an alias, delete its name below in the same change.

allowed_aliases=(
)

actual="$(
  (rg --no-filename --only-matching --replace '$1' '^type ([A-Za-z0-9_]+) =' \
    internal/port --glob '*.go' --glob '!**/*_test.go' --max-depth 1 || true) | sort -u
)"
expected="$(printf '%s\n' "${allowed_aliases[@]}" | sort -u)"

if [[ "$actual" != "$expected" ]]; then
  printf '%s\n' 'internal/port type aliases changed; do not add aliases, and keep the allowlist in sync:' >&2
  diff -u <(printf '%s\n' "$expected") <(printf '%s\n' "$actual") >&2 || true
  exit 1
fi

printf '%s\n' 'port alias ratchet check passed'
