-- mcp_tool_invocation_id sengaja UUID polos — bagian dari siklus tiga
-- arah yang sama dengan komentar di 000017. Ditutup di 000020.
CREATE TABLE hitl_approvals (
    id                          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    mcp_tool_invocation_id      UUID,
    requested_at                TIMESTAMPTZ NOT NULL DEFAULT now(),
    approver_user_id            UUID REFERENCES users(id) ON DELETE SET NULL,
    approval_channel            TEXT NOT NULL CHECK (approval_channel IN ('librechat_ui','rest_api','cli')),
    decision                    TEXT NOT NULL DEFAULT 'pending' CHECK (decision IN ('pending','approved','rejected','expired')),
    decided_at                  TIMESTAMPTZ,
    decision_reason             TEXT,
    original_command_snapshot   JSONB NOT NULL
);
