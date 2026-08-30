# Polyglot Agent Instructions

## Source Of Truth

- Read this file first; `CLAUDE.md` points here and must not duplicate it.
- Follow `DEVELOPMENT-GUIDELINES.md` for implementation rules.
- Follow `EFFECTIVE_GO.md` for Go writing, formatting, naming, errors, and package design.
- Use `Polyglot-Architecture.md` for architecture rationale and `PLAN-BACKEND-STANDARDIZATION-V2.md` for remaining work.
- `opencode.jsonc` configures the repository `code-review-graph` MCP server; use graph tools before filesystem exploration when available.

## Repository Shape

- Root is one Go module, targeting Go `1.26`; frontend is separate under `web/`.
- Runtime entrypoints: `cmd/server`, `cmd/probe`, and `cmd/seed`.
- `internal/domain/`: pure models, invariants, value objects, and domain errors. No transport, port, adapter, driver, database, or framework imports.
- `internal/usecase/`: orchestration over domain and `internal/port`; never import `internal/adapter` or `internal/driver`.
- `internal/port/`: consumer-owned interfaces and contracts, not business entities.
- `internal/adapter/connect/<context>/`: ConnectRPC handlers, routers, and the only protobuf/domain mappers for that context.
- `internal/adapter/http/`: HTTP handlers and middleware; use `net/http.ServeMux` and `middleware.Chain`.
- `internal/driver/<vendor>/`: external hardware/protocol implementations.
- `api/gen/` and `web/src/gen/` are generated; never edit manually.
- `test/integration/` uses explicit `integration` build tags and service dependencies.

## Non-Negotiable Rules

- Map transport errors through `pkg/response`; handlers must not construct ad hoc `connect.NewError` values.
- Create domain sentinels with `fault.New` in the owning domain package; wrap propagated errors with `%w`.
- Use `pkg/logger`; production code must not use `log.Printf`, `log.Println`, or `fmt.Println`.
- Production logs use static messages, snake_case fields, request correlation ID, and no secrets, tokens, cookies, full payloads, or sensitive PII.
- Context is a parameter, never a struct field. Every goroutine has cancellation or wait ownership.
- Keep production files under 500 lines; mark deliberate exceptions `// DEVIASI: <reason>`.
- Preserve protobuf field names, numbers, services, and wire semantics. Add fields only backward-compatibly.

## Commands

```text
make build
make vet
make test
make lint
make check
make proto-check
make check-connect-errors check-layer-boundaries
make test-integration
make security
```

- `make test` runs `go test ./... -race -cover`.
- `make test-integration` runs tagged tests in `test/integration`; start PostgreSQL/Redis with `make db-up` or provide `TEST_DATABASE_URL` and required integration environment.
- `make test-mikrotik-e2e` requires a reachable MikroTik router and explicit `MIKROTIK_TEST_*` configuration.
- Proto changes require `buf lint`, generation through `make proto-check`, and a clean generated diff.
- Run focused tests with `go test ./path/to/package -run TestName -count=1`.
- Run `gofmt -w <changed-go-files>` after Go edits.

## Completion

- Before claiming completion, run relevant focused tests, then `make build`, `make vet`, `make test`, `make lint`, and both boundary checks.
- Run `make test-integration` for integration changes and `make security` for dependency/security changes.
- AI-authored commits must include a `Co-Authored-By` trailer identifying agent model and attribution.
