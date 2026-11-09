# ADR 0011: Single-Tenant Deployment

## Status
Accepted

## Context
The codebase carries tenant identifiers from an earlier multi-tenant design.
Every runtime path uses the same value, database defaults include
`tenant-default`, and protobuf messages already expose tenant fields. Keeping
those fields is necessary for wire and schema compatibility, but maintaining a
tenant abstraction with no second tenant adds indirection, unsafe defaults, and
untested code paths.

## Decision
Polyglot runs as a single-tenant deployment.

- `internal/domain/tenant.DefaultID` is the only application-level source of
  the tenant identifier (`tenant-default`).
- Application code must not introduce additional tenant IDs or accept a tenant
  as user input. New code uses `tenant.DefaultID` at the boundary.
- Existing database columns, defaults, and protobuf fields remain in place for
  backward compatibility. They are not an invitation to add multi-tenant
  behavior.
- Reintroducing multi-tenancy requires a new ADR, schema design, authorization
  model, and cross-tenant isolation tests.

## Consequences
Tenant literals are checked by `scripts/check-tenant-literals.sh`. The legacy
wire and database surface stays stable, while new code has one obvious,
testable default. The main trade-off is that tenant-scoped columns remain
redundant; that is preferable to a risky schema migration that provides no
current product value.
