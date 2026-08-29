#!/usr/bin/env bash
set -euo pipefail

if rg -n 'connect\.NewError' internal/adapter/connect --glob '*.go'; then
  printf '%s\n' 'direct connect.NewError is forbidden in ConnectRPC handlers'
  exit 1
fi

printf '%s\n' 'ConnectRPC error boundary check passed'
