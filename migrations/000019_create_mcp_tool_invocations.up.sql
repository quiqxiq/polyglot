-- conversation_id sengaja UUID polos — ai_conversations belum ada
-- sampai 000020. Ditutup di situ. command_audit_log_id di bawah AMAN
-- inline (bukan ditunda) karena command_audit_log sudah ada sejak 000017.
CREATE TABLE mcp_tool_invocations (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    conversation_id         UUID,
    tool_name               TEXT NOT NULL,
    input_params            JSONB,
    actor_type              TEXT NOT NULL CHECK (actor_type IN ('ai_agent','human_direct','scheduled_job')),
    requires_approval       BOOLEAN NOT NULL DEFAULT false,
    approval_status         TEXT NOT NULL DEFAULT 'not_required' CHECK (approval_status IN ('not_required','pending','approved','rejected')),
    command_audit_log_id    UUID REFERENCES command_audit_log(id) ON DELETE SET NULL,
    invoked_at              TIMESTAMPTZ NOT NULL DEFAULT now()
);
