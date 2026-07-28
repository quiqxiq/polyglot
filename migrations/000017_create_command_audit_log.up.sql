-- Satu tabel audit untuk SEMUA eksekusi command ke device, siapa pun
-- pelakunya (staff manual, AI agent lewat MCP, atau job terjadwal).
-- classification & decision PERSIS mencerminkan command.Class dan
-- command.Decision di internal/domain/command — lihat
-- DATABASE-SCHEMA.md §9.1.
--
-- hitl_approval_id sengaja UUID polos (bukan REFERENCES) — ini bagian
-- dari siklus tiga arah command_audit_log <-> hitl_approvals <->
-- mcp_tool_invocations yang tidak bisa dibuat inline sekaligus. FK ini
-- ditutup di 000020_create_ai_conversations.up.sql, migration TERAKHIR
-- di rantai siklus ini. Lihat DATABASE-SCHEMA.md §9.1 catatan ⚠.
CREATE TABLE command_audit_log (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    device_id           UUID NOT NULL REFERENCES devices(id) ON DELETE RESTRICT,
    actor_type          TEXT NOT NULL CHECK (actor_type IN ('human','ai_agent','system_scheduled')),
    actor_user_id       UUID REFERENCES users(id) ON DELETE SET NULL,
    actor_display_name  TEXT,
    source              TEXT NOT NULL CHECK (source IN ('mcp_tool','rest_api','scheduled_job','manual_cli')),
    command_raw         TEXT NOT NULL,
    command_args        JSONB,
    classification      TEXT NOT NULL CHECK (classification IN ('read_only','destructive')),
    decision            TEXT NOT NULL CHECK (decision IN ('auto_approved','required_approval','denied')),
    hitl_approval_id    UUID,
    executed_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    completed_at        TIMESTAMPTZ,
    success             BOOLEAN,
    result_summary      TEXT,
    error_message       TEXT
);

-- Menutup FK yang ditunda dari 000011_create_provisioning_sync_log —
-- command_audit_log sekarang sudah ada.
ALTER TABLE provisioning_sync_log
    ADD CONSTRAINT fk_sync_log_command_audit
    FOREIGN KEY (command_audit_log_id) REFERENCES command_audit_log(id) ON DELETE SET NULL;
