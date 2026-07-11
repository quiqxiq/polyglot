CREATE TABLE ai_conversations (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    actor_user_id   UUID REFERENCES users(id) ON DELETE SET NULL,
    channel         TEXT NOT NULL CHECK (channel IN ('claude_desktop','web_chat','mobile','cli')),
    started_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    ended_at        TIMESTAMPTZ,
    summary         TEXT
);

CREATE TABLE ai_messages (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    conversation_id UUID NOT NULL REFERENCES ai_conversations(id) ON DELETE CASCADE,
    role            TEXT NOT NULL CHECK (role IN ('user','assistant','tool')),
    content         TEXT NOT NULL,
    model_name      TEXT,
    tokens_used     INTEGER,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Penutup siklus tiga arah: ai_conversations sekarang sudah ada, jadi
-- ketiga FK yang ditunda dari 000017/000018/000019 dilengkapi di sini,
-- urutan sesuai dependensi. Lihat DATABASE-SCHEMA.md §9.4.
ALTER TABLE mcp_tool_invocations
    ADD CONSTRAINT fk_mcp_invocation_conversation
    FOREIGN KEY (conversation_id) REFERENCES ai_conversations(id) ON DELETE SET NULL;

ALTER TABLE hitl_approvals
    ADD CONSTRAINT fk_hitl_mcp_invocation
    FOREIGN KEY (mcp_tool_invocation_id) REFERENCES mcp_tool_invocations(id) ON DELETE CASCADE;

ALTER TABLE command_audit_log
    ADD CONSTRAINT fk_command_audit_hitl
    FOREIGN KEY (hitl_approval_id) REFERENCES hitl_approvals(id) ON DELETE SET NULL;
