# ADR 0011 — AI Router Ops: Approval + Restore Point

Status: Accepted (31 Oktober 2026)

## Context

AI/MCP tools could execute RouterOS mutations without server-side approval
(`ExecuteCommandPreApproved`) and `/mcp` was unauthenticated. RouterOS API has
no "commit confirmed", so a bad mutation can lock the operator out.

## Decision

1. Commands are classified as **READ** or **MUTATION** (unknown = MUTATION).
   READ executes; MUTATION only creates an approval request.
2. Every MUTATION requires a restore point bundle (scoped export, binary
   backup, inverse script) stored encrypted by Polyglot before approval exists.
3. Human approval is server-side via ConnectRPC with JWT + RBAC; AI/MCP cannot
   approve, and `/mcp` is JWT-protected with full command audit.
4. Execution is one-shot (row lock + revision), post-verified, and audit-logged.
5. Rollback: inverse script + binary restore are primary; the router-scheduler
   watchdog is optional (`WATCHDOG_ENABLED=false` default) until verified on CHR.
6. `/console/inspect` adds pre-approval syntax validation on RouterOS v7; v6
   runs in degraded mode with warnings and never crashes.
7. The active router for AI is a server-side binding (`active_devices`), not a
   `.env` address or client-only state.

## Consequences

- Legacy `run_command`/`push_config` now only propose mutations.
- Mikhmon mutating tools are not exposed to AI v1 (no raw-command diff model).
- Every mutation pays for a backup; dedupe/retention jobs bound growth.
- Restore requires typed device-name confirmation and is UI-only.
