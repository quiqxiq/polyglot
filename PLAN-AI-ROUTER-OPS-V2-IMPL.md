# PLAN-AI-ROUTER-OPS v2 — Implementation Record

Status: **implemented (M0–M5 core)**. This record accompanies the approved replan
and supersedes stale assumptions in [`integrasi-router-ai.md`](integrasi-router-ai.md).

## Invariants

- READ commands auto-execute with audit.
- Every MUTATION requires: validation → diff → restore point → human approval.
- AI/MCP can never approve its own request; approval is only via human JWT.
- Restore point missing/failed = proposal rejected (fail-closed).
- Restore requires typed device-name confirmation and RBAC `restore:apply`.
- No hidden chain-of-thought is rendered; only structured events.
- AI hard-deny: `reset-configuration`, `backup/load`, AI-created scripts/schedulers,
  arbitrary `/file/remove` and `/tool/fetch`. Reboot remains a normal mutation.

## Milestones completed

| Milestone | Isi |
|---|---|
| M0 | `command.Mode` READ/MUTATION + `Classification`/`DenyReason`; classifier baru semua driver; hapus `ExecuteCommandPreApproved`; `/mcp` JWT+RBAC+audit+actor; hapus tool Mikhmon mutating dari MCP; config + RBAC baru |
| M1 | migrate `000041` (restore_points/artifacts, approval_requests, command_logs, active_devices, ai_conversations, ai_chat_events); storage terenkripsi AES-GCM; driver restore-point (export/binary/read/cleanup); diff & inverse; proposal/decision/execution usecase; ApprovalService + RestorePointService |
| M2 | UI `/ai-ops/approvals` (diff, edit, approve/reject, expiry/revision), `/ai-ops/restore-points` (list/detail/download/restore typed confirmation) |
| M3 | `/console/inspect` client (child/syntax/completion/highlight), capability v7/v6 + cache, MCP read/inspect/capabilities tools; active-device binding (Redis/Postgres, `Set/Get/ClearActiveDevice`) + device-switcher badge |
| M4 | `AIChatService.Chat` server-streaming; chat agent tools (read/inspect/propose/approval/restore metadata); event persistence; chat UI (`/ai-ops`) |
| M5 | approval expiry job, restore cleanup retention, daily/weekly scheduled backups, optional watchdog (default OFF) via router scheduler |

## Known limitations (v1)

- Chat text is delivered as a single `TextDelta` after the tool loop (Genkit
  streaming-tool spike not completed); tool/plan events stream live.
- Binary restore upload is limited to ~60KB over the API; larger backups must be
  restored via manual file transfer + `/system/backup/load`.
- Explicit restore-point deletion is implemented (soft delete + encrypted file
  cleanup; refuses bundles referenced by active approvals).
- Retention jobs: approval expiry, restore-point age/keep-last cleanup,
  command_logs and AI chat event retention are implemented.
- Binary backup dedupe uses `ConfigHash` + storage hardlink (copy fallback);
  critical menus (`/system`, `/ip/firewall`, `/ip/route`, `/interface`,
  `/ip/service`, `/ipv6`) intentionally always get a fresh binary backup.
- MCP auth v1 uses Bearer JWT; static API tokens/OAuth are deferred.

## Verification

See the final verification section in the session record: `go build/vet`,
`go test` focused packages, migration smoke (`TESTCONTAINERS_RYUK_DISABLED=true`),
`buf lint` + `buf generate` idempotent, boundary checks, frontend `tsc`,
`vite build`, and ESLint.

## Files

- Migration: `migrations/000041_ai_router_ops.{up,down}.sql`
- Domains: `internal/domain/{command,approval,restorepoint,aichat}`
- Drivers: `internal/driver/mikrotik/{classify.go,ai_ops.go,inspect/}`
- Usecases: `internal/usecase/{network,aichat,device}`
- Adapters: `internal/adapter/{postgres,storage,connect/{approval,restorepoint,aichat},mcp}`
- Web: `web/src/features/ai-ops`, `web/src/routes/_authenticated/ai-ops`
