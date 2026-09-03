#!/usr/bin/env bash
set -euo pipefail

if rg -n 'connect\.NewError' internal cmd --glob '*.go'; then
  printf '%s\n' 'direct connect.NewError is forbidden in internal and cmd packages; map through pkg/response'
  exit 1
fi

printf '%s\n' 'ConnectRPC error boundary check passed'
