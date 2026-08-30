# ADR 0006: HTTP Error Envelope

## Status
Accepted

## Decision
Plain HTTP endpoints return JSON errors with `error.code`, `error.message`, and optional safe `error.details`. `pkg/response` owns mapping from `fault.Kind` to status and public code. Internal error text, stack traces, credentials, and database details never cross this boundary.

## Consequence
Clients can branch on stable machine-readable codes. Existing endpoint migrations must preserve compatibility or use an explicit versioned contract.
