#!/usr/bin/env bash
set -euo pipefail

if rg -n '"github\.com/quixiq/polyglot/internal/(adapter|driver)' internal/usecase --glob '*.go' --glob '!**/*_test.go'; then
  printf '%s\n' 'usecase production code must not import adapter or driver'
  exit 1
fi

if rg -n '"github\.com/quixiq/polyglot/(internal/(port|usecase|adapter|driver)|api/gen)' internal/domain --glob '*.go'; then
  printf '%s\n' 'domain code must not import transport, port, usecase, adapter, or driver packages'
  exit 1
fi

printf '%s\n' 'layer boundary check passed'
